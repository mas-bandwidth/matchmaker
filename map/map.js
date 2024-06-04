
const dt = 1.0 / 60.0
const standard_width = 2000
const standard_height = 940
const width = 120
const height = 64
const size = width * height

const origin_x = 0
const origin_y = 0

const spacing_x = 16
const spacing_y = 16

const color_radius = 4
const max_radius = 7

const background = "rgb(10,10,10)"

;(function () {
  let canvas, ctx, raw_mouse_x, raw_mouse_y, data, fadeout

  function init() {

    canvas = document.getElementById('gameCanvas')

    ctx = canvas.getContext('2d')

    window.requestAnimationFrame(update)

    t = 0.0

    data = Array.apply(null, Array(size)).map(function (x, i) { return 0.0; }) 

    fadeout = Array.apply(null, Array(size)).map(function (x, i) { return 0.0; }) 

    raw_mouse_x = -1000
    raw_mouse_y = -1000

    canvas.addEventListener('mousemove', (e) => {
      raw_mouse_x = e.offsetX
      raw_mouse_y = e.offsetY
    })

    canvas.addEventListener('mouseleave', (e) => {
      raw_mouse_x = -1000
      raw_mouse_y = -1000
    })

    update_data()

    setInterval(update_data, 10)
  }

  function update_data() {
    new_data = load_binary_resource("http://127.0.0.1:8000/data")
    if (new_data.length == size*4) {
      let buffer = new ArrayBuffer(new_data.length)
      let view8 = new Uint8Array(buffer)
      for (var i = 0; i < new_data.length; i++) {
        view8[i] = new_data[i]
      }
      let view32 = new Uint32Array(buffer)
      for (var i = 0; i < size; i++) {
        data[i] += ( view32[i] - data[i] ) * 0.01
      }
    }
  }

  function load_binary_resource(url) {
    var byteArray = []
    try {
      var req = new XMLHttpRequest()
      req.open('GET', url, false)
      req.overrideMimeType('text\/plain; charset=x-user-defined')
      req.send(null)
    } catch(error) {
      return byteArray
    }

    if (req.status != 200) return byteArray
    for (var i = 0; i < req.responseText.length; ++i) {
      byteArray.push(req.responseText.charCodeAt(i) & 0xff)
    }
    return byteArray
  }

  function update() {

    window.requestAnimationFrame(update)

    ctx.rect(0, 0, canvas.width, canvas.height)
    ctx.fillStyle = background
    ctx.fill()

    canvas_width = canvas.getBoundingClientRect().width

    normalize_factor = canvas_width / standard_width

    mouse_x = raw_mouse_x / normalize_factor
    mouse_y = raw_mouse_y / normalize_factor

    var radius = 0.0

    for (var j = 0; j < height-15; j++) {
      for (var i = 0; i < width; i++) {

        index = i + j*width

        draw = false

        if (data[index] > 0.00001) {
          fadeout[index] += ( 1.0 - fadeout[index] ) * 0.99999
        } else {
          fadeout[index] *= 0.99999
        }

        if (fadeout[index] > 0.00001) {
          draw = true
          intensity = data[index] / 70
          r = 50 * (0.25 + 0.25 * intensity)
          g = 200 * (0.25 + 0.25 * intensity)
          b = 255 * (0.25 + 0.25 * intensity)
          r = 10 + fadeout[index] * r
          g = 10 + fadeout[index] * g
          b = 10 + fadeout[index] * b
          color = 'rgb(' + r + ',' + g + ',' + b + ')'
          intensity2 = intensity
          intensity2 -= 15
          if (intensity2 < 0) {
            intensity2 = 0
          }
          radius = color_radius + intensity2 * 2
          if (radius > max_radius) {
            radius = max_radius
          }
        }
      
        // draw circle

        x = origin_x + i*spacing_x
        y = origin_y + j*spacing_y

        x *= normalize_factor
        y *= normalize_factor

        radius *= normalize_factor

        if (draw) {
          ctx.fillStyle = color
          ctx.beginPath()
          ctx.arc(x, y, radius, 0, 2 * Math.PI, true)
          ctx.fill()
          ctx.closePath()
        }
      }
    }

    t += dt
  }

  document.addEventListener('DOMContentLoaded', init)
})()
