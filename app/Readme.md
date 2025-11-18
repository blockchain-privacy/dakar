# Frontend

This is the frontend of Dakar. It gives access generic blockchain data like transactions, blocks and addresses.
Additionally, it serves as an analytics platform for Dash private send transactions.

## Dependencies

* `vue v3` - frontend framework
* `vuetify v3` - vue component library
* `d3` -  chart and graph library

For a more detailed overview check [here](./package.json).


## Getting started

* Install [yarn version 4](https://github.com/yarnpkg/berry) or higher
* To use private javascript packages, write the following content in your `.yarnrc.yml` located in your home folder:

```yaml
yarnPath: .yarn/releases/yarn-4.0.0.cjs # set right path/version here

npmScopes:
  blockchain:
    npmRegistryServer: "https://<gitlab_host>/api/v4/projects/410/packages/npm/"
    npmAlwaysAuth: true
    npmAuthToken: "<your-deploy-token>"
```

## Front-end setup

```shell
yarn install
```

### Compiles and hot-reloads for development
```shell
yarn dev
```

### Compiles and minifies for production
```shell
yarn build
```

### Starts a local server using the files produced by `yarn build`
```shell
yarn preview
```

### Update project

```shell
yarn up '*'
```



