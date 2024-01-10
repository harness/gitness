module github.com/drone/drone

replace github.com/docker/docker => github.com/docker/engine v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible

require (
	github.com/99designs/httpsignatures-go v0.0.0-20170731043157-88528bf4ca7e
	github.com/Azure/azure-storage-blob-go v0.7.0
	github.com/asaskevich/govalidator v0.0.0-20180315120708-ccb8e960c48f
	github.com/aws/aws-sdk-go v1.37.3
	github.com/coreos/go-semver v0.2.0
	github.com/dchest/authcookie v0.0.0-20120917135355-fbdef6e99866
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/drone/drone-go v1.7.2-0.20220308165842-f9e4fe31c2af
	github.com/drone/drone-runtime v1.1.1-0.20200623162453-61e33e2cab5d
	github.com/drone/drone-ui v2.12.0+incompatible
	github.com/drone/drone-yaml v1.2.4-0.20220204000225-01fb17858c9b
	github.com/drone/envsubst v1.0.3-0.20200709231038-aa43e1c1a629
	github.com/drone/funcmap v0.0.0-20210823160631-9e9dec149056
	github.com/drone/go-license v1.0.2
	github.com/drone/go-login v1.1.0
	github.com/drone/go-scm v1.28.0
	github.com/drone/signal v1.0.0
	github.com/dustin/go-humanize v1.0.0
	github.com/go-chi/chi v1.5.5
	github.com/go-chi/chi/v5 v5.0.11
	github.com/go-chi/cors v1.0.0
	github.com/go-redis/redis/v8 v8.11.0
	github.com/go-redsync/redsync/v4 v4.3.0
	github.com/go-sql-driver/mysql v1.4.0
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.9
	github.com/google/go-jsonnet v0.17.0
	github.com/google/wire v0.2.1
	github.com/gosimple/slug v1.3.0
	github.com/h2non/gock v1.0.15
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-retryablehttp v0.5.4
	github.com/hashicorp/golang-lru v0.5.1
	github.com/jmoiron/sqlx v0.0.0-20180614180643-0dae4fefe7c0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/lib/pq v1.1.0
	github.com/mattn/go-sqlite3 v1.9.0
	github.com/oxtoacart/bpool v0.0.0-20150712133111-4e1c5567d7c2
	github.com/prometheus/client_golang v1.16.0
	github.com/robfig/cron v0.0.0-20180505203441-b41be1df6967
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.6.0
	github.com/unrolled/secure v0.0.0-20181022170031-4b6b7cf51606
	github.com/yarlson/chiprom v0.1.0
	go.starlark.net v0.0.0-20221020143700-22309ac47eac
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/sync v0.3.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Azure/azure-pipeline-go v0.2.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Microsoft/go-winio v0.4.11 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar v1.1.1 // indirect
	github.com/buildkite/yaml v2.1.0+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/codegangsta/negroni v1.0.0 // indirect
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/go-connections v0.3.0 // indirect
	github.com/docker/go-units v0.3.3 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gogo/protobuf v0.0.0-20170307180453-100ba4e88506 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/h2non/parth v0.0.0-20190131123155-b4df798d6542 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/mattn/go-ieproxy v0.0.0-20190610004146-91bb50d98149 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/natessilva/dag v0.0.0-20180124060714-7194b8dcc5c4 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be // indirect
	github.com/vinzenz/yaml v0.0.0-20170920082545-91409cdd725d // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/grpc v1.30.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.14

go 1.21
