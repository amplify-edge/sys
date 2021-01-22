module github.com/getcouragenow/sys

go 1.15

require (
	github.com/Masterminds/squirrel v1.4.0
	github.com/VictoriaMetrics/metrics v1.12.3
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/genjidb/genji v0.9.0
	github.com/genjidb/genji/engine/badgerengine v0.9.0
	github.com/getcouragenow/sys-share v0.0.0-20201211115435-35b645d047ee
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/improbable-eng/grpc-web v0.13.0
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rs/cors v1.7.0 // indirect
	github.com/segmentio/encoding v0.2.2
	github.com/sendgrid/rest v2.6.2+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.7.1+incompatible
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20201116153603-4be66e5b6582
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	google.golang.org/grpc v1.33.2
	google.golang.org/grpc/examples v0.0.0-20201117005946-20636e76a99a // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/getcouragenow/sys-share => ../sys-share/
