# Spamhaus wrapper

### How to run

```docker compose up --build```

### How to use
- go to the GraphQL playground http://localhost:3030/ uses basic auth username and password like specified in requirements.
- use mutation to add IP addresses for checking: 
```graphql
mutation {
    updateIPDetails(ips: ["127.0.0.2", "127.0.0.4", "127.0.0.10", "31.5.99.86"]) {
        uuid
        ipAddress
        responseCode
        createdAt
        updatedAt
    }
}
```
- use query to actually see the mutated data.
```graphql
query {
    getIPDetails(ip: "31.5.99.86") {
        uuid
        ipAddress
        responseCode
        createdAt
        updatedAt
    }
}
```

### How to test
- run unit tests with ```go test ./...```

#### Notes regarding test implementation:
- We're using a mock resolver to avoid actual DNS queries during tests

### Libraries
- uuid from github.com/google/uuid -> to handle UUID types
- gqlgen from github.com/99designs/gqlgen/graphql -> for building GraphQL servers in Go.
- gqlparser to parse the graphql schema files
- github.com/stretchr/testify for testing. Using assert and require from this package for clearer test assertions.
- github.com/DATA-DOG/go-sqlmock for testing. To generate mocks in repository.
