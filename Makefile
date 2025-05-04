up: down
	docker compose up --build -d

down:
	docker compose down

clean:
	docker compose down -v

lint:
	make -C search-services lint

proto:
	make -C search-services protobuf
