module.exports = {
  devServer: {
    proxy: "http://localhost:8081",
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": 'GET, POST',
      "Access-Control-Allow-Headers": "Origin, Content-Type, X-Auth-Token",
    }
  },

  transpileDependencies: [
    "vuetify",
  ]
}