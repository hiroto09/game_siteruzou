up:
	docker compose build --no-cache && docker-compose up -d 
down:
	docker compose down
restart:
	docker compose restart
logs:
	docker compose logs -f
login:
	docker exec -it db mysql -u root -p

# MySQL初期化のみ行う（DB立ち上げ後に手動で schema.sql を流す）
init-db:
	docker compose up -d --build &&\
	@echo "⏳ DB起動を待機中..." &&\
	sleep 20  &&\
	docker exec -i db mysql -u root -p$${MYSQL_ROOT_PASSWORD} app < ./db/schema.sql &&\
	@echo "✅ DBスキーマ初期化完了"
