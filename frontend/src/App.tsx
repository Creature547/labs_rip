import { useEffect, useMemo, useState } from 'react'
import { Button, Card, Col, Container, Form, Navbar, Row } from 'react-bootstrap'
import defaultImage from './assets/default-equipment.png'
import defaultVideo from './assets/default-video.svg'
import './App.css'

type Service = {
  id: number
  name: string
  description: string
  power: number
  cost: number
  createdAt: string
  imagePath?: string
  videoPath?: string
}

const mockServices: Service[] = [
  {
    id: 1,
    name: 'Цилиндрическая печь',
    description: 'Современная цилиндрическая печь-камин для эффективного обогрева загородного дома.',
    power: 12.5,
    cost: 99900,
    createdAt: '2026-09-06',
    imagePath: '/media/images/cylindrical-stove.png',
    videoPath: '/media/videos/cylindrical-stove.mp4',
  },
  {
    id: 2,
    name: 'Бойлер косвенного нагрева',
    description: 'Высокопроизводительный бойлер для стабильной подачи горячей воды и экономичной работы.',
    power: 24,
    cost: 65000,
    createdAt: '2026-09-08',
    imagePath: '/media/images/indirect-boiler.png',
    videoPath: '/media/videos/indirect-boiler.mp4',
  },
  {
    id: 3,
    name: 'Электрическая печь',
    description: 'Электрокамин с эффектом живого пламени и быстрым безопасным обогревом жилой комнаты.',
    power: 3,
    cost: 35000,
    createdAt: '2026-09-10',
    imagePath: '/media/images/electric-stove.png',
    videoPath: '/media/videos/electric-stove.mp4',
  },
  {
    id: 4,
    name: 'Теплогенератор',
    description: 'Промышленный теплогенератор высокой мощности для обогрева больших помещений и производственных зон.',
    power: 200,
    cost: 200000,
    createdAt: '2026-09-12',
    imagePath: '/media/images/heat-generator.png',
    videoPath: '/media/videos/heat-generator.mp4',
  },
]

const price = (cost: number) =>
  new Intl.NumberFormat('ru-RU').format(cost) + ' ₽'

function TileImage({ service }: { service: Service }) {
  const [failed, setFailed] = useState(false)

  return (
    <Card.Img
      className="tile-image"
      variant="top"
      src={!failed && service.imagePath ? service.imagePath : defaultImage}
      alt={service.name}
      onError={() => setFailed(true)}
    />
  )
}

function FeedMedia({ service }: { service: Service }) {
  const [videoFailed, setVideoFailed] = useState(false)
  const [imageFailed, setImageFailed] = useState(false)

  if (!service.videoPath || videoFailed) {
    return (
      <img
        className="feed-media"
        src={!imageFailed && service.imagePath ? service.imagePath : defaultVideo}
        alt={service.name}
        onError={() => setImageFailed(true)}
      />
    )
  }

  return (
    <video
      className="feed-media"
      src={service.videoPath}
      poster={service.imagePath || defaultVideo}
      autoPlay
      muted
      loop
      playsInline
      controls
      onError={() => setVideoFailed(true)}
    />
  )
}

function TileCard({ service }: { service: Service }) {
  return (
    <Card className="tile-card h-100">
      <TileImage service={service} />

      <Card.Body className="d-flex flex-column">
        <Card.Title>{service.name}</Card.Title>

        <Card.Text>{service.description}</Card.Text>

        <div className="mt-auto d-flex justify-content-between align-items-center">
          <strong>{price(service.cost)}</strong>
          <small>
            {new Date(service.createdAt).toLocaleDateString('ru-RU')}
          </small>
        </div>
      </Card.Body>
    </Card>
  )
}

function FeedCard({
  service,
  onNext,
}: {
  service: Service
  onNext: () => void
}) {
  const [expanded, setExpanded] = useState(false)

  const short =
    service.description.length > 72
      ? service.description.slice(0, 72) + '…'
      : service.description

  return (
    <section className="feed-card">
      <FeedMedia service={service} />
      <div className="feed-shade" />

      <div className="feed-content">
        <div className="feed-heading">
          <h1>{service.name}</h1>
        </div>

        <div className="feed-info">
          <span>
            Мощность
            <b>{service.power} кВт</b>
          </span>

          <span>
            Стоимость
            <b>{price(service.cost)}</b>
          </span>
        </div>

        <p>
          {expanded ? service.description : short}{' '}
          {service.description.length > 72 && (
            <button onClick={() => setExpanded(!expanded)}>
              {expanded ? 'свернуть' : 'ещё'}
            </button>
          )}
        </p>

        <Button
          className="next-round"
          onClick={onNext}
          aria-label="Следующая услуга"
        >
          →
        </Button>
      </div>
    </section>
  )
}

function App() {
  const [page, setPage] = useState<'feed' | 'tiles'>('feed')
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState('new')
  const [feedIndex, setFeedIndex] = useState(0)

  const filtered = useMemo(() => {
    const found = mockServices.filter((service) =>
      service.name.toLowerCase().includes(query.trim().toLowerCase()),
    )

    return [...found].sort((a, b) =>
      sort === 'price'
        ? a.cost - b.cost
        : +new Date(b.createdAt) - +new Date(a.createdAt),
    )
  }, [query, sort])

  useEffect(() => {
    setFeedIndex(0)
  }, [page])

  return (
    <div className="app-shell">
      <main>
        {page === 'feed' ? (
          <Container className="feed-container">
            <FeedCard
              service={mockServices[feedIndex]}
              onNext={() =>
                setFeedIndex((index) => (index + 1) % mockServices.length)
              }
            />
          </Container>
        ) : (
          <Container className="tiles-container">
            <h1>Плитка услуг</h1>

            <p className="tiles-note">
              Выберите отопительное оборудование для своего дома.
            </p>

            <Row className="g-3 filter-row">
              <Col md={7}>
                <Form.Control
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Поиск по названию"
                />
              </Col>

              <Col md={5}>
                <Form.Select
                  value={sort}
                  onChange={(event) => setSort(event.target.value)}
                >
                  <option value="new">Сначала новые</option>
                  <option value="price">По возрастанию цены</option>
                </Form.Select>
              </Col>
            </Row>

            <p className="found">Найдено услуг: {filtered.length}</p>

            <Row className="g-4">
              {filtered.map((service) => (
                <Col key={service.id} sm={6} lg={3}>
                  <TileCard service={service} />
                </Col>
              ))}
            </Row>
          </Container>
        )}
      </main>

      <Navbar fixed="bottom" className="bottom-nav">
        <Container className="justify-content-around px-0">
          <Button
            variant="link"
            className={page === 'feed' ? 'nav-item active' : 'nav-item'}
            onClick={() => setPage('feed')}
          >
            ◉
            <span>Лента</span>
          </Button>

          <Button variant="link" className="nav-item disabled" disabled>
            ＋
            <span>Добавление</span>
          </Button>

          <Button
            variant="link"
            className={page === 'tiles' ? 'nav-item active' : 'nav-item'}
            onClick={() => setPage('tiles')}
          >
            ▦
            <span>Плитка</span>
          </Button>

          <Button variant="link" className="nav-item disabled" disabled>
            ⇥
            <span>Вход</span>
          </Button>
        </Container>
      </Navbar>
    </div>
  )
}

export default App