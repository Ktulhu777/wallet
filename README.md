<h1>API Документация</h1>

<h2>🚀 Используемые технологии</h2>
<ul>
  <li><img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"></li>
  <li><img src="https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL"></li>
  <li><img src="https://img.shields.io/badge/Docker-24.0-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"></li>
</ul>

<h2>📌 Установка проекта</h2>
<pre>
  git clone https://github.com/Ktulhu777/wallet.git
  cd wallet
  docker-compose up --build
</pre>

<h2>📌 API Эндпоинты</h2>
<table>
  <thead>
    <tr>
      <th>Метод</th>
      <th>Эндпоинт</th>
      <th>Описание</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>GET</td>
      <td>/api/v1/wallets/{WALLET_UUID}</td>
      <td>Получение информации о кошельке</td>
    </tr>
    <tr>
      <td>POST</td>
      <td>/api/v1/wallet</td>
      <td>Создание или изменение кошелька</td>
    </tr>
  </tbody>
</table>

<h2>📌 Пример запроса</h2>
<pre>
  curl -X GET http://127.0.0.1:7777/api/v1/wallets/550e8400-e29b-41d4-a716-446655440000
</pre>
