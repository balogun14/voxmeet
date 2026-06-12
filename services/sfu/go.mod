module github.com/awwal/voxmeet/sfu

go 1.25.5

require (
	github.com/awwal/voxmeet/pkgs v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rabbitmq/amqp091-go v1.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/awwal/voxmeet/pkgs => ../../pkgs
