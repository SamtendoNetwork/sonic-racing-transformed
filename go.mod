module github.com/nex-server/sonic-transformed

go 1.24.0

require (
	github.com/PretendoNetwork/grpc/go v0.0.0-20260501210425-981c793afb28
	github.com/PretendoNetwork/nex-go/v2 v2.3.1
	github.com/PretendoNetwork/nex-protocols-common-go/v2 v2.6.1
	github.com/PretendoNetwork/nex-protocols-go/v2 v2.3.7
	github.com/PretendoNetwork/plogger-go v1.1.0
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/beorn7/perks v1.0.1
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/dolthub/maphash v0.1.0
	github.com/fatih/color v1.18.0
	github.com/jwalton/go-supportscolor v1.2.0
	github.com/klauspost/compress v1.18.0
	github.com/lxzan/gws v1.8.8
	github.com/mattn/go-colorable v0.1.14
	github.com/mattn/go-isatty v0.0.20
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.66.1
	github.com/prometheus/procfs v0.16.1
	github.com/rasky/go-lzo v0.0.0-20200203143853-96a758eda86e
	github.com/superwhiskers/crunch/v3 v3.5.7
	go.yaml.in/yaml/v2 v2.4.2
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0
	golang.org/x/mod v0.29.0
	golang.org/x/net v0.47.0
	golang.org/x/sys v0.38.0
	golang.org/x/term v0.37.0
	golang.org/x/text v0.31.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda
)

replace github.com/PretendoNetwork/nex-go/v2 => ./third_party/nex-go

replace github.com/PretendoNetwork/nex-protocols-common-go/v2 => ./third_party/nex-protocols-common-go-trace
