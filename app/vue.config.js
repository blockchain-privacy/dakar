// webpack is part of vue-cli, DO NOT add it as dependency. It is going to result in several errors.
// eslint-disable-next-line import/no-extraneous-dependencies
const webpack = require('webpack');
const childProcess = require('child_process');

const gitCommitHash = childProcess.execSync('git rev-parse --short HEAD').toString();
const gitBranch = childProcess.execSync('git branch --show-current').toString();

module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
      },
      '/auth': {
        target: 'http://localhost:4433',
        // remove '/auth' prefix
        pathRewrite: { '^/auth': '' },
      },
      '/wikiapi': {
        target: 'http://localhost:4567',
        // remove '/wiki' prefix
        pathRewrite: { '^/wikiapi': '' },
      },
    },
  },
  chainWebpack: (config) => {
    config
      .plugin('html')
      .tap((args) => {
        // better use applicationName in constants/index.js,
        // but the node js server only understands common.js. We use ES6.
        args[0].title = 'Dakar';
        return args;
      });
  },
  transpileDependencies: [
    'vuetify',
  ],
  runtimeCompiler: true,
  configureWebpack: {
    plugins: [
      new webpack.DefinePlugin({
        __COMMIT_HASH__: JSON.stringify(gitCommitHash.trim()),
      }),
      new webpack.DefinePlugin({
        __BRANCH__: JSON.stringify(gitBranch.trim()),
      }),
    ],
  },
};
