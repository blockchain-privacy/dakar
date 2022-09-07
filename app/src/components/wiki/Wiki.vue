<template>
  <div>
    <p>Hello,World</p>
  </div>
</template>

<script>
import { PAGE_TITLE } from '../../constants';

export default {
  name: 'Wiki',
  mounted() {
    document.title = `Wiki - ${PAGE_TITLE}`;
  },
};
</script>

<style scoped>

</style>
