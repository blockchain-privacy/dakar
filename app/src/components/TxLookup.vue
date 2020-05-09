<template>

  <v-layout row>
    <v-flex xs12 sm8 offset-sm2>
      <v-card>
        <v-card-title>
          <v-icon>mdi-bank-transfer</v-icon>
          Transaction
        </v-card-title>
        <v-list two-line subheader>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-format-header-pound</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>HASH</v-list-item-title>
              <v-list-item-subtitle>{{data.tx.hash}}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon>mdi-calendar</v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>DATE</v-list-item-title>
              <v-list-item-subtitle>ts: {{data.tx.bts}} - {{new Date(data.tx.bts)}}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Confirmations</v-list-item-title>
              <v-list-item-subtitle>{{data.confirmations}}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Block</v-list-item-title>
              <v-list-item-subtitle>{{data.bheight}} - {{data.bhash}}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Inputs</v-list-item-title>
              <v-list-item-subtitle>{{data.tx.inputs.length}}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-avatar>
              <v-icon></v-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>Outputs</v-list-item-title>
              <v-list-item-subtitle>{{data.tx.outputs.length}}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-list>
      </v-card>
      <v-card><!-- TODO hack, to leave space at the bottom, such that content is not hidden by the footer.-->
        <div style="height: 40pt">
        </div>
      </v-card>
    </v-flex>
  </v-layout>

  <!-- example tx json struct
   {"tx":
     {"hash":"a6f50fba49f976dc4cfc8536f2ad4aec9ee09edc56b8ef7c15046f3582d72c56",
      "inputs":[
         {"txhash":"0a0d8017a9d7b2b32f0e5d10a1319e1495866c44303670391e6295609551db11",
          "txtype":"",
          "amount":15.247015,
          "addresses":["XoHFzfbFHos7cumFgDhfWCWubRNYTL9xoA"],
          "index":0,
          "iscoinbase":false},
         {"txhash":"1abb74cca8cc42297a62130b7b2749f102313265c785b058a472d63723f4c6f3","txtype":"","amount":100,"addresses":["XiwvtUe8g6BLjvLkiVkWqgd8URwuniRcZD"],"index":1,"iscoinbase":false},{"txhash":"41a7de03b772dd3645933af4f9fbaf905ea168d5bbd83d8943470afcb57b765c","txtype":"","amount":100,"addresses":["XiwvtUe8g6BLjvLkiVkWqgd8URwuniRcZD"],"index":2,"iscoinbase":false},{"txhash":"48be033e19ae4273e68b3dff153535b937b156522d5a57f09da78957c597f5f9","txtype":"","amount":100,"addresses":["XiwvtUe8g6BLjvLkiVkWqgd8URwuniRcZD"],"index":3,"iscoinbase":false},{"txhash":"c24489969ef3efee1ae05646b56f1d97ffd8c4dcafca8b42ffa167496822ddbc","txtype":"","amount":100,"addresses":["XiwvtUe8g6BLjvLkiVkWqgd8URwuniRcZD"],"index":4,"iscoinbase":false},{"txhash":"f51d6377b177a66a98d2e4eb94c9205aa7cffc0093d29c02e96cfdbee1cff98f","txtype":"","amount":100,"addresses":["XiwvtUe8g6BLjvLkiVkWqgd8URwuniRcZD"],"index":5,"iscoinbase":false},{"txhash":"f60f0188b51d5c1ccb37fb18f2f65f3398118a9738408bada294dd218e68acc4","txtype":"","amount":100,"addresses":["XiwvtUe8g6BLjvLkiVkWqgd8URwuniRcZD"],"index":6,"iscoinbase":false}],
      "outputs":[
         {"txhash":"a6f50fba49f976dc4cfc8536f2ad4aec9ee09edc56b8ef7c15046f3582d72c56",
          "txtype":"pubkeyhash",
          "amount":1.83946,
          "addresses":["XvpW92QQCQkfbjJJaZZmebUTG9Cxdba9Hf"],
          "index":0,
          "iscoinbase":false},
         {"txhash":"a6f50fba49f976dc4cfc8536f2ad4aec9ee09edc56b8ef7c15046f3582d72c56","txtype":"pubkeyhash","amount":613.402,"addresses":["Xrxj471GS44z5T6vABf3syocnV9Hef3zsy"],"index":1,"iscoinbase":false}],"ts":1534206756},
      "bhash":"00000000000000341469d7ab5aa190cbf49a19ac69afcf8cfd608d7f8cdf7245",
      "bheight":920000,
      "bts":1534206756,
      "confirmations":168640,
      "version":1}
    -->

</template>

<script>
  export default {
    name: 'TxLookup',
    props: ['data']
  }
</script>
