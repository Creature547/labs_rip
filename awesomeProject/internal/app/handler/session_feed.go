package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"awesomeProject/internal/app/currentuser"
	"awesomeProject/internal/app/model"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/serializer"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const feedBatchSize = 100

var sessionIDPattern = regexp.MustCompile("^[a-f0-9]{32}$")

// Each batch is written atomically so concurrent requests see the same order.
var saveFeedBatch = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then
  for i, id in ipairs(ARGV) do
    redis.call('ZADD', KEYS[1], i, id)
  end
end
redis.call('EXPIRE', KEYS[1], 3600)
return redis.call('ZRANGE', KEYS[1], 0, -1)
`)

type feedRepository interface {
	RandomPublishedIDs(context.Context, int, []uint) ([]uint, error)
	GetPublishedByID(uint) (*model.HeatingEquipment, error)
	MediaBucket() string
}

type SessionFeed struct {
	repo         feedRepository
	redis        *redis.Client
	mediaBaseURL string
}

type SessionFeedCard struct {
	Item     serializer.HeatingEquipment `json:"item"`
	Position int                         `json:"position"`
}

func NewSessionFeed(repo feedRepository, redisClient *redis.Client, mediaBaseURL string) *SessionFeed {
	return &SessionFeed{repo: repo, redis: redisClient, mediaBaseURL: mediaBaseURL}
}

// Get godoc
// @Summary Карточка из случайной ленты текущей сессии
// @Tags Приборы
// @Produce json
// @Param session query string true "Случайный идентификатор из 32 hex-символов"
// @Param position query integer false "Сквозной порядковый номер карточки" default(1)
// @Success 200 {object} SessionFeedCard
// @Failure 400 "Неверные параметры"
// @Failure 404 "Приборы отсутствуют"
// @Failure 503 "Redis недоступен"
// @Router /heating-equipment/session-feed [get]
func (h *SessionFeed) Get(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	position, err := strconv.Atoi(c.DefaultQuery("position", "1"))
	session := c.Query("session")
	if err != nil || position < 1 || position > 1000000 || !sessionIDPattern.MatchString(session) {
		c.Status(http.StatusBadRequest)
		return
	}

	userID := currentuser.ID(c)
	prefix := fmt.Sprintf("heating:feed:%d:%s:", userID, session)
	ctx := c.Request.Context()
	firstBatch := (position-1)/feedBatchSize + 1
	for batch := firstBatch; batch <= firstBatch+1; batch++ {
		ids, err := h.loadBatch(ctx, prefix, batch)
		if err != nil {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		if len(ids) == 0 {
			break
		}
		start := max(0, position-1-(batch-1)*feedBatchSize)
		for index := start; index < len(ids); index++ {
			id, err := strconv.ParseUint(ids[index], 10, 64)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			item, err := h.repo.GetPublishedByID(uint(id))
			if errors.Is(err, repository.ErrNotFound) {
				continue
			}
			if err != nil {
				handleRepositoryError(c, err)
				return
			}
			c.JSON(http.StatusOK, SessionFeedCard{
				Item:     serializer.HeatingEquipmentFromModel(*item, userID, h.mediaBaseURL, h.repo.MediaBucket()),
				Position: (batch-1)*feedBatchSize + index + 1,
			})
			return
		}
	}
	c.Status(http.StatusNotFound)
}

func (h *SessionFeed) loadBatch(ctx context.Context, prefix string, batch int) ([]string, error) {
	key := prefix + strconv.Itoa(batch)
	ids, err := h.redis.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		return ids, h.redis.Expire(ctx, key, time.Hour).Err()
	}

	var exclude []uint
	if batch > 1 {
		previous, err := h.redis.ZRange(ctx, prefix+strconv.Itoa(batch-1), 0, -1).Result()
		if err != nil {
			return nil, err
		}
		for _, id := range previous {
			parsed, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				return nil, err
			}
			exclude = append(exclude, uint(parsed))
		}
	}
	generated, err := h.repo.RandomPublishedIDs(ctx, feedBatchSize, exclude)
	if err != nil {
		return nil, err
	}
	if len(generated) < feedBatchSize && len(exclude) > 0 {
		generated, err = h.repo.RandomPublishedIDs(ctx, feedBatchSize, nil)
		if err != nil {
			return nil, err
		}
	}
	if len(generated) == 0 {
		return nil, nil
	}
	args := make([]any, 0, len(generated))
	for _, id := range generated {
		args = append(args, strconv.FormatUint(uint64(id), 10))
	}
	return saveFeedBatch.Run(ctx, h.redis, []string{key}, args...).StringSlice()
}
