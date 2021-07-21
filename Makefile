.PHONY: all
all: down monitoring storage simulator

.PHONY: logs
logs:
	docker-compose logs --tail 100 -f $(SERVICE)

.PHONY: tail-topic
tail-topic:
	docker-compose run redpanda -- "rpk topic --brokers redpanda:9092 consume $(TOPIC)"

.PHONY: monitoring
monitoring:
	docker-compose rm -fsv prometheus grafana
	docker-compose up -d prometheus grafana

.PHONY: storage
storage:
	docker-compose rm -fsv singlestore redpanda
	docker-compose up -d singlestore redpanda

.PHONY: simulator
simulator:
	docker-compose rm -fsv simulator
	docker-compose up --build -d simulator

.PHONY: down
down:
	docker-compose down -v