# Serve Swagger UI

## Updating the Swagger UI assets

1. Copy all assets from [Swagger-ui repo](https://github.com/swagger-api/swagger-ui/tree/master/dist) to `swagger-ui` directory.
2. Rename `index.html` to `index.gothml`
3. In `index.gohtml` set `url` of `SwaggerUIBundle` to `{{ .specPath }}`

Done 👍

Now execute tests and make sure they pass. 