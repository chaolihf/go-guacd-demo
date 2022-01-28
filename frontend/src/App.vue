<template>
  <div id="app">
    <remote-terminal/>
  </div>
</template>

<script>
import RemoteTerminal from "@/views/remote";

export default {
  name: 'app',
  components: {
    RemoteTerminal,
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

html,body,#app {
  padding: 0px;
  margin: 0px;
  height: 100%;
}
</style>
