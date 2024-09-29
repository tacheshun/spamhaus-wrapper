# Spamhaus wrapper

### How to run
- docker compose up -d --build

### Libraries
- uuid from github.com/google/uuid -> to handle UUID types
- gqlgen from github.com/99designs/gqlgen/graphql -> for building GraphQL servers in Go. 
It simplifies schema definition, code generation, and resolver implementation
- gqlparser to parse the graphql schema files

### Issues

When testing the code for the challange, I was always getting error response from the zen spamhaus API.

Tried setting DNS to 1.1.1.1, also tried leave it on auto to get the ISP resolver DNS, but I got the same result.

➜  ~ nslookup 2.0.0.127.zen.spamhaus.org 1.1.1.1
Server:		1.1.1.1
Address:	1.1.1.1#53

Non-authoritative answer:
Name:	2.0.0.127.zen.spamhaus.org
Address: 127.255.255.254


Tried also 8.8.8.8 and 8.8.4.4 I was getting:
** server can't find 2.0.0.127.zen.spamhaus.org: NXDOMAIN

