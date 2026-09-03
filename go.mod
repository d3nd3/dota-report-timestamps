module github.com/d3nd3/dota-report-timestamps

go 1.25.0

require (
	github.com/dotabuff/manta v1.5.0
	github.com/golang/protobuf v1.5.4
	github.com/klauspost/compress v1.20.0
	github.com/paralin/go-dota2 v0.0.0-20250623204622-532cb1851217
	github.com/paralin/go-steam v0.0.0-20250502043548-f167ff28a93a
	github.com/shurcooL/graphql v0.0.0-20240915155400-7ee5256398cf
	github.com/sirupsen/logrus v1.10.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.47.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/paralin/go-steam => github.com/paralin/go-steam v0.0.0-20250502043548-f167ff28a93a

replace github.com/paralin/go-dota2 => github.com/paralin/go-dota2 v0.0.0-20250623204622-532cb1851217
