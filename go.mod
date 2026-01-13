module github.com/APTrust/dart-runner

go 1.24.11

// Force use of nsq v1.2.0 because v1.3.0 has breaking changes.
replace github.com/nsqio/nsq => github.com/nsqio/nsq v1.2.0

require (
	github.com/APTrust/preservation-services v0.0.0-20250121215436-e111cbb1f03f
	github.com/brianvoe/gofakeit/v6 v6.28.0
	github.com/dimchansky/utfbom v1.1.1
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/minio/minio-go/v7 v7.0.84
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/pkg/sftp v1.13.7
	github.com/stretchr/stew v0.0.0-20130812190256-80ef0842b48b
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.47.0
	golang.org/x/text v0.33.0
	modernc.org/sqlite v1.34.5
)

require (
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bmizerany/perks v0.0.0-20230307044200-03f9df79da1e // indirect
	github.com/bytedance/sonic v1.12.7 // indirect
	github.com/bytedance/sonic/loader v0.2.3 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.0.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.24.0 // indirect
	github.com/go-redis/redis/v7 v7.4.1 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/nsqio/go-diskqueue v0.0.0-20180306152900-74cfbc9de839 // indirect
	github.com/nsqio/go-nsq v1.1.0 // indirect
	github.com/nsqio/nsq v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.13.0 // indirect
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.61.9 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.2 // indirect
)
