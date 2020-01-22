# sabadoscodes.com frontend

## Project setup
```
npm install
```
### Compiles and hot-reloads for development
```
make run
```
Note, builds need an SSM parameter entered for the google analytics id to use. See the See 
[the infrastructure README.md](../infrastructure/README.md) for more. 
### Compiles and minifies for production
```
make
```
Note, builds need an SSM parameter entered for the google analytics id to use. See the See 
[the infrastructure README.md](../infrastructure/README.md) for more. 
### Run your unit tests
```
make test
```

### Lints and fixes files
```
npm run lint
```

### Deploying
```
./bucket_sync.sh
```
Note, `bucket_sync.sh` requires the AWS cli be setup with credentials appropriate to your account as well as the proper
entries in SSM. See [the infrastructure README.md](../infrastructure/README.md) for more.

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).
