<template>
<div>
    <svg width="500" height="500">
      <g
          class="flower"
          v-for="flower in layoutData.children"
          :key="flower.data.name"
          :style="{
          transform: `translate(${flower.x}px, ${flower.y}px)`
        }"
      >
        <circle class="flower__circle" :r="flower.data.amount" :fill="flower.data.color"></circle>
        <text class="flower__label">{{ flower.data.name }}</text>
      </g>
    </svg>
    <div class="controls">
      <div class="control" v-for="flower in flowers" :key="flower.name">
        <label>{{ flower.name }}</label>
        <input type="number" v-model="flower.amount" step="10" min="10">
      </div>
    </div>
</div>
</template>

<script>
import * as d3 from "d3";
export default {
  name: "HeuristicEditor",
  data() {
    return {
      flowers: [
        {
          name: 'Roses',
          amount: 25,
          color: '#cc2936'
        },
        {
          name: 'Tulips',
          amount: 40,
          color: '#f2c640'
        },
        {
          name: 'Daisies',
          amount: 15,
          color: '#2a93d4'
        },
        {
          name: 'Narcissuses',
          amount: 9,
          color: '#F7AD0A'
        }
      ]
    }
  },
  computed: {
    transformedFlowerData() {
      return {
        name: 'Top Level',
        children: this.flowers.map(flower => ({
          ...flower,
          parent: 'Top Level'
        }))
      }
    },

    layoutData() {
      // Generate a D3 hierarchy
      const rootHierarchy =
          d3.hierarchy(this.transformedFlowerData)
              .sum(d => d.size)
              .sort((a, b) => {
                return b.value - a.value
              })
      // Pack the circles inside the viewbox
      return d3.pack()
          .size([500, 500])
          .padding(10)(rootHierarchy)
    }
  },
  created() {
    this.$store.dispatch('updateHeuristicData',this.$route.params.id);
  },
}
</script>

<style scoped>
svg {
  display: block;
  margin: 0 auto;
}

.flower {
  transition: transform 0.2s ease-in-out;
  text-anchor: middle;
}

.flower__circle {
  transition: r 0.2s ease-in-out;
}

.flower__label {
  fill: #fff;
  font-weight: bold;
  text-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
}

.controls {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.control {
  display: inline-flex;
  flex-direction: column;
  margin: 0 4px;
}

.control label {
  font-size: 14px;
  font-weight: bold;
  margin-bottom: 4px;
}

.control input {
  display: block;
  font: inherit;
  width: 100px;
}
</style>