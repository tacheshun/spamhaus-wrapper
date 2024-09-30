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

### Issues encountered

1. When testing the code for the challenge, I was always getting error response from the spamhaus Zend API. 
I don't know the cause...it might be because my IP seems to be blacklisted. 

Tried setting DNS to 1.1.1.1, also tried leave it on auto to get the ISP resolver DNS, but I got the same result.

➜  ~ nslookup 2.0.0.127.zen.spamhaus.org 1.1.1.1
Server:		1.1.1.1
Address:	1.1.1.1#53

Non-authoritative answer:
Name:	2.0.0.127.zen.spamhaus.org
Address: 127.255.255.254

Tried also 8.8.8.8 and 8.8.4.4 I was getting:
** server can't find 2.0.0.127.zen.spamhaus.org: NXDOMAIN

2. Made the createdAt and updatedAt nullable.