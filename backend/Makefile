all: run

run:
	rm log
	DB_HOST=localhost go run ./src/cmd | tee -a log

connect-db:
	docker exec -it languasia-db psql languasia languasia

