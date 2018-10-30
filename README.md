# go-eveonline
Go package for Eve Online ESI and SDE.

This package is in active development, I am filling in the pieces as I need them.  Pull requests are very welcome!  If you are using this library and want a particular feature implemented, open an issue and I will try and get to it.

# Usage

```
go get github.com/pequalsnp/go-eveonline
```

Then in your code

```
import (
  "github.com/pequalsnp/go-eveonline/pkg/esi"
  "github.com/pequalsnp/go-eveonline/pkg/eveonline"
)
```

# Authorized Requests

The pattern this library uses is compatible with `golang.org/x/oauth2`.  This library allows you to use the `Client(ctx, token)` function to get a client that will automatically set the access token header properly and handles refreses when needed.  I *strongly* recommend you use it.

Whenever a method takes an `authdClient *http.Client` it assunmes you will pass in such a client built with a token with the required scopes.

# Examples

TBD - I know they would be helpful and I will try and add some to this repo.
