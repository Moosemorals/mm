
# The code to run my website

## To run (in dev mode)

    cd main
    go run main.go -debug :8080,:8443

Explanation:

  `-debug` uses local certs (hard coded as cert.pem and key.pem)

  `:8080,:8443` is a pair of addresses (in 'ip:port' format, either side is
  optional, but must have at least one) to listen on. The first address
  is http and redirects to the second address, which is https.

  You can add as many address pairs as you need.
