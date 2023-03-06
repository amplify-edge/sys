module go.amplifyedge.org/sys-v2

go 1.16

replace go.amplifyedge.org/sys-share-v2 => ../sys-share/

require (
	github.com/Masterminds/squirrel v1.5.0
	github.com/VictoriaMetrics/metrics v1.13.0
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/genjidb/genji v0.10.1
	github.com/genjidb/genji/engine/badgerengine v0.10.0
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/improbable-eng/grpc-web v0.14.0
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/cors v1.7.0 // indirect
	github.com/segmentio/encoding v0.2.7
	github.com/sendgrid/rest v2.6.2+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.7.2+incompatible
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	go.amplifyedge.org/sys-share-v2 v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.1.0
	golang.org/x/net v0.1.0
	google.golang.org/grpc v1.35.0
	google.golang.org/grpc/examples v0.0.0-20210205041354-b753f4903c1b // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.6 // indirect
)
