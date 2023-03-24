<template>
  <el-container>
    <el-main>
      <div id='viewport' ref="viewport" class="viewport"/>
    </el-main>
    <el-aside>
      <h2 style="text-align: center; margin-top: 20rem; font-size: 18px">远程桌面（RDP/VNC/SSH）</h2>
      <el-form label-position="center" style="margin: 2rem 1rem 1rem 0" label-width="80px">
        <el-form-item label="资产">
          <el-select v-model="query.remote" placeholder="请选择要连接的资产" @change="doSelectChange"
                     :disabled="connected">
            <el-option label="RDP" value="rdp-server"/>
            <el-option label="VNC" value="vnc-server"/>
            <el-option label="SSH" value="ssh-server"/>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width:45%" @submit="doConnect" @click="doConnect" :disabled="connected">
            连接
          </el-button>
          <el-button style="width: 45%" @click="doDisconnect" :disabled="!connected">断开</el-button>
        </el-form-item>
      </el-form>
    </el-aside>
  </el-container>
</template>

<script>
import Guacamole from 'guacamole-common-js'

export default {
  name: "remote-terminal",
  data() {
    return {
      client: null,
      mouse: null,
      keyboard: null,
      connected: false,
      query: {
        remote: 'rdp-server',
        width: 1024,
        height: 768,
        port: 3389
      }
    }
  },
  computed: {
    wsUrl() {
      return "ws://127.0.0.1:9528/ws"
    }
  },
  methods: {
    doConnect() {
      const viewport = this.$refs.viewport
      if (!viewport || !viewport.offsetWidth) {
        return        // resize is being called on the hidden window
      }

      // console.log(viewport.clientWidth)
      this.query.width = viewport.clientWidth
      this.query.height = viewport.clientHeight
      console.log('viewport:', viewport.clientWidth, 'x', viewport.clientHeight)

      let tunnel = new Guacamole.WebSocketTunnel(this.wsUrl)

      tunnel.onerror = error => {
        this.$alert(`tunnel failed: ${JSON.stringify(error)}`)
      }
      tunnel.onstatechange = state => {
        switch (state) {
          case Guacamole.Tunnel.State.OPEN:
            console.log("switch to state_open")
            setTimeout(() => {
              this.resize()
              // this.client.sendSize(viewport.clientWidth, viewport.clientHeight)


              const element = this.client.getDisplay().getElement();
              this.$refs.viewport.appendChild(element)

              // FIXME: Set all canvas to the front
              const canvas = this.$refs.viewport.getElementsByTagName('canvas')
              for (let c of canvas) {
                c.setAttribute('style', 'z-index: 99999;');
              }
            }, 1000)
            break
          case Guacamole.Tunnel.State.CONNECTING:
            console.log('switch to state_connecting')
            break
          case Guacamole.Tunnel.State.CLOSED:
            console.log('switch to state_closed')
            break
          case Guacamole.Tunnel.State.UNSTABLE:
            console.log('switch to state_unstable')
            break
        }
      }

      if (this.client) {
        this.doDisconnect()
      }
      this.client = new Guacamole.Client(tunnel)
      this.client.onerror = error => {
        this.$alert(error)
      }

      // mouse
      this.mouse = new Guacamole.Mouse(this.client.getDisplay().getElement())
      this.mouse.onmousedown = this.mouse.onmouseup = this.mouse.onmousemove = (mouseState) => {
        this.client.sendMouseState(mouseState)
      }
      // keyboard
      if (!this.keyboard) {
        this.keyboard = new Guacamole.Keyboard(document)
      }
      this.keyboard.onkeydown = (keysym) => {
        this.client.sendKeyEvent(1, keysym)
      }
      this.keyboard.onkeyup = (keysym) => {
        this.client.sendKeyEvent(0, keysym)
      }
      this.client.connect(this.serialize(this.query))
      this.connected = true

      window.addEventListener('resize', this.resize)

    },
    doDisconnect() {
      if (this.client) {
        this.client.getDisplay().scale(0)
        this.client.disconnect()
      }
      if (this.mouse) {
        this.mouse.onmousedown = this.mouse.onmouseup = this.mouse.onmousemove = null
      }
      if (this.keyboard) {
        this.keyboard.onkeydown = this.keyboard.onkeyup = null
      }
      this.client = null
      this.mouse = null
      this.connected = false
    },
    serialize(query) {
      let str = []
      for (const p in query) {
        if (query[p]) {
          str.push(encodeURIComponent(p) + '=' + encodeURIComponent(query[p]))
        }
      }
      return str.join('&')
    },
    doSelectChange(event) {
      switch (event) {
        case 'rdp-server':
          this.query.port = 3389
          break
        case 'vnc-server':
          this.query.port = 5901
          break
        case 'ssh-server':
          this.query.port = 22
          break
      }
    },
    resize() {
      if (this.connected) {
        const viewport = this.$refs.viewport
        if (!viewport || !viewport.offsetWidth) {
          console.log('viewport is null')
          console.log(viewport)
          return        // resize is being called on the hidden window
        }
        // this.client.resize(viewport.clientWidth, viewport.clientHeight)

        this.client.sendSize(viewport.clientWidth, viewport.clientHeight)
        console.log("resize:", viewport.clientWidth, 'x', viewport.clientHeight)
      }
    }
  }
}
</script>

<style scoped>
.el-container {
  height: 100%;
}

.el-main {
  background-color: indianred;
  margin: 0;
  padding: 0;
  overflow: hidden;
}

.el-aside {
  width: 300px;
  min-width: 200px;
  background-color: wheat;
}

.viewport {
  width: 100%;
  height: 100%;
  overflow: hidden;
  cursor: none;
}

</style>