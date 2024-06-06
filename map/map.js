
const dt = 1.0 / 60.0
const standard_width = 1900
const standard_height = 940
const width = 120
const height = 64
const size = width * height

const origin_x = 0
const origin_y = 0

const spacing_x = 16
const spacing_y = 16

const min_radius = 2
const mid_radius = 5
const max_radius = 7.5

const background = "rgb(15,15,15)"

;(function () {
  let canvas, ctx, data, fadeout, max_players

  function init() {

    canvas = document.getElementById('gameCanvas')

    ctx = canvas.getContext('2d')

    window.requestAnimationFrame(update)

    t = 0.0

    max_players = 0.0

    data = Array.apply(null, Array(size)).map(function (x, i) { return 0.0; }) 

    radius = Array.apply(null, Array(size)).map(function (x, i) { return min_radius; }) 

    fadeout = Array.apply(null, Array(size)).map(function (x, i) { return 0.0; }) 

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

    new_max_players = 0.0

    for (var j = 0; j < height; j++) {
      for (var i = 0; i < width; i++) {

        index = i + j*width

        draw = false

        if (data[index] > new_max_players) {
          new_max_players = data[index]
        }

        if (data[index] > 0.00001) {
          fadeout[index] += ( 1.0 - fadeout[index] ) * 0.01
        }

        var rad = min_radius

        if (fadeout[index] > 0.00001) {

          draw = true

          m = max_players / 10

          intensity = data[index] / m
          r = 25 + 75 * (0.25 + intensity)
          g = 25 + 170 * (0.25 + intensity)
          b = 25 + 255 * (0.25 + intensity)
          r = 15 + fadeout[index] * r
          g = 15 + fadeout[index] * g
          b = 15 + fadeout[index] * b
          color = 'rgb(' + r + ',' + g + ',' + b + ')'

          if (data[index] < m ) {
            rad += ( mid_radius - min_radius ) * ( data[index] / m )
          } else {
            d = data[index] - m
            rad += ( max_radius - mid_radius ) * ( d / m )
          }
          if (rad > max_radius) {
            rad = max_radius
          }
        }

        radius[index] += ( rad - radius[index] ) * 0.25
      
        rad = radius[index]

        // draw circle

        x = origin_x + i*spacing_x
        y = origin_y + j*spacing_y

        x *= normalize_factor
        y *= normalize_factor

        rad *= normalize_factor

        if (draw) {
          ctx.fillStyle = color
          ctx.beginPath()
          ctx.arc(x, y, rad, 0, 2 * Math.PI, true)
          ctx.fill()
          ctx.closePath()
        }
      }
    }

    max_players = new_max_players

    t += dt
  }

  document.addEventListener('DOMContentLoaded', init)
})()
