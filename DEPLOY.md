# Deploy Finora lên VPS

Repo có sẵn cấu hình production tại `docker-compose.production.yml`. Cấu hình này dùng PostgreSQL Finora đã cài trên VPS qua Docker network `finora-net`; API không mở cổng ra Internet. Website chạy ở cổng `2001` và tự proxy `/api` vào backend.

Trên VPS, chạy một lần sau khi clone repo:

```bash
git clone https://github.com/mapmapdicode/Finora.git /opt/finora/app
cd /opt/finora/app
cp .env.production.example .env.production
chmod 600 .env.production
nano .env.production
```

Trong `.env.production`, thay các giá trị `CHANGE_ME`, đặc biệt là `APP_JWT_SECRET` và SMTP. Tạo JWT secret bằng:

```bash
openssl rand -hex 48
```

Kiểm tra database network đã tồn tại, sau đó build và chạy:

```bash
docker network inspect finora-net
docker compose -f docker-compose.production.yml up -d --build
docker compose -f docker-compose.production.yml ps
curl http://localhost:2001
```

Nếu VPS dùng UFW, mở website:

```bash
ufw allow 2001/tcp
```

Truy cập `http://110.172.29.117:2001`. Khi cập nhật mã nguồn:

```bash
cd /opt/finora/app
git pull
docker compose -f docker-compose.production.yml up -d --build
```

Lưu ý: file `/opt/finora/postgres.env` phải vẫn tồn tại trên VPS và có `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`. File này là secret của database, không đưa vào Git. Sau khi website deploy ổn định, nên giới hạn hoặc đóng cổng PostgreSQL `5433` nếu không còn cần kết nối từ máy local.
