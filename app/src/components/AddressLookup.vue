<template>
    <v-layout row>
        <v-flex xs12 sm8 offset-sm2>
            <v-card>
                <v-card-title>
                    <v-icon>mdi-bank-transfer</v-icon>
                    Address
                </v-card-title>
                <v-list two-line subheader>
                    <v-list-item>
                        <v-list-item-avatar>
                            <v-icon>mdi-format-header-pound</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>HASH</v-list-item-title>
                            <v-list-item-subtitle>{{data.address}}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item v-for="tx in data.txs" v-bind:key="tx.txhash">
                        <v-list-item-avatar>
                            <v-icon></v-icon>
                        </v-list-item-avatar>
                        <v-list-item-content>
                            <v-list-item-title>Transaction</v-list-item-title>
                            <v-list-item-subtitle>Hash: {{tx.txhash}}</v-list-item-subtitle>
                            <v-list-item-subtitle>Amount: {{tx.amount}}</v-list-item-subtitle>
                            <v-list-item-subtitle>Index: {{tx.index}}</v-list-item-subtitle>
                            <v-list-item-subtitle v-if="tx.iscoinbase" :data="tx.iscoinbase">Coinbase:
                                {{tx.iscoinbase}}
                            </v-list-item-subtitle>
                            <v-list-item-subtitle v-if="tx.txtype" :data="tx.txtype">Type: {{tx.txtype}}
                            </v-list-item-subtitle>
                            <div v-if="tx.addresses.length > 1">
                                <v-list-item-subtitle v-for="addr in tx.addresses" v-bind:key="addr">Addresses:
                                    {{addr}}
                                </v-list-item-subtitle>
                            </div>
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
      {
      "address": "XpUHYnvS2d4NsC2GqULvne4iqFd1QuhNzi",
      "clusters": null,
      "txs": [
        {
          "txhash": "0fde0403cd951be6883930d14229300fe0a05a8c9857d0a9ebc2a141817a2977",
          "txtype": "",
          "amount": 0.03257228,
          "addresses": [
            "XpUHYnvS2d4NsC2GqULvne4iqFd1QuhNzi"
          ],
          "index": 0,
          "iscoinbase": false
        }
      ]
     }
     -->
</template>

<script>
    export default {
        name: 'AddressLookup',
        props: ['data']
    }
</script>
