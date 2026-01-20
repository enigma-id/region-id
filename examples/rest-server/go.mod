module rest-server

go 1.24.4

replace github.com/enigma-id/region-id => ../../

require (
	github.com/enigma-id/region-id v0.1.2
	github.com/joho/godotenv v1.5.1
	github.com/logistics-id/engine v0.0.19-dev
	github.com/logistics-id/engine/ds/redis v0.0.19-dev
	github.com/logistics-id/engine/transport/rest v0.0.19-dev
	github.com/uptrace/bun v1.2.16
	github.com/uptrace/bun/dialect/pgdialect v1.2.16
	github.com/uptrace/bun/driver/pgdriver v1.2.16
	github.com/uptrace/bun/extra/bundebug v1.2.16
	go.uber.org/zap v1.27.1
)

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/gomodule/redigo v1.9.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/logistics-id/engine/common v0.0.19-dev // indirect
	github.com/logistics-id/engine/validate v0.0.19-dev // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	mellium.im/sasl v0.3.2 // indirect
)
