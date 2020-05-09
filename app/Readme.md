# DashRPC front-end app

This is the front-end of the Dash explorer.

## Getting started

* Build the front-end
* Build `crawler` on the backend project
* Run `crawler` for a small-test block range, such that badger DB is generated
* Build `explorer` on the backend project
* Run Dash daemon (`dashd`)
* Run `explorer`
* Deploy the front-end app
* Seach for TX from the blocks that you have cached in badger


## Front-end setup
```
yarn install
```

### Compiles and hot-reloads for development
```
yarn serve
```

### Compiles and minifies for production
```
yarn build
```

### Lints and fixes files
```
yarn lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).


