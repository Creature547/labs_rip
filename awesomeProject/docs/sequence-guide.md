# Sequence для StarUML: сценарий ЛР4

Это инструкция для построения, не готовая нативная диаграмма.

Участники: пользователь (Actor), SPA (Lifeline), UserDomain (Lifeline),
EquipmentDomain (Lifeline). БД, Redis, MinIO и нативное приложение на этой
диаграмме не показываются согласно заданию.

В диаграмме классов домены должны содержать операции. В StarUML задайте типы
lifeline доменными классами. При создании сообщения выберите соответствующую
операцию в поле signature. Текст arguments храните отдельно — не заменяйте
signature просто написанной строкой названия метода.

| № | SPA → домен, signature | arguments входа | Домен → SPA: ключевой результат |
|---|---|---|---|
| 1 | UserDomain.Authenticate | login, password | access_token, user, expires_in |
| 2 | EquipmentDomain.List | min_power=1, max_cost=10000 | items[] |
| 3 | EquipmentDomain.CreateDraft | JWT, name, description, power, cost, image, video | id, поля черновика, image_url, video_url |
| 4 | EquipmentDomain.Draft | JWT | id, поля своего черновика |
| 5 | EquipmentDomain.Publish | JWT, id | status=опубликован, formed_at |
| 6 | EquipmentDomain.Feed | без id | первый опубликованный прибор, id |
| 7 | EquipmentDomain.FeedByID | id, next=true | следующий опубликованный прибор, id |
| 8 | EquipmentDomain.Like | JWT, id, like=1 | likes_count, liked=1 |

Сопоставление HTTP адресов находится в README. Возвратные сообщения должны
иметь тип reply, выходные данные — в arguments. Шаги пользователя к SPA можно
назвать «Войти», «Задать фильтр», «Добавить прибор», «Открыть черновик»,
«Опубликовать», «Открыть ленту», «Следующий», «Поставить лайк».

Для скриншота 9 выделите сообщение Publish: в свойствах signature должна быть
именно ссылка на EquipmentDomain.Publish из диаграммы классов.
