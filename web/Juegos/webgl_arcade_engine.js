(function (global) {
  function createRenderer(canvas) {
    var gl = canvas.getContext("webgl", { antialias: true, alpha: false });
    if (!gl) {
      throw new Error("WebGL no disponible");
    }

    var vertexSource =
      "attribute vec3 aPosition;" +
      "attribute vec3 aColor;" +
      "uniform mat4 uMatrix;" +
      "varying vec3 vColor;" +
      "void main(){ gl_Position = uMatrix * vec4(aPosition, 1.0); vColor = aColor; }";
    var fragmentSource =
      "precision mediump float;" +
      "varying vec3 vColor;" +
      "void main(){ gl_FragColor = vec4(vColor, 1.0); }";

    function shader(type, source) {
      var s = gl.createShader(type);
      gl.shaderSource(s, source);
      gl.compileShader(s);
      if (!gl.getShaderParameter(s, gl.COMPILE_STATUS)) {
        throw new Error(gl.getShaderInfoLog(s));
      }
      return s;
    }

    var program = gl.createProgram();
    gl.attachShader(program, shader(gl.VERTEX_SHADER, vertexSource));
    gl.attachShader(program, shader(gl.FRAGMENT_SHADER, fragmentSource));
    gl.linkProgram(program);
    if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
      throw new Error(gl.getProgramInfoLog(program));
    }
    gl.useProgram(program);

    var posLoc = gl.getAttribLocation(program, "aPosition");
    var colorLoc = gl.getAttribLocation(program, "aColor");
    var matrixLoc = gl.getUniformLocation(program, "uMatrix");
    var buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.enableVertexAttribArray(posLoc);
    gl.enableVertexAttribArray(colorLoc);
    gl.vertexAttribPointer(posLoc, 3, gl.FLOAT, false, 24, 0);
    gl.vertexAttribPointer(colorLoc, 3, gl.FLOAT, false, 24, 12);
    gl.enable(gl.DEPTH_TEST);

    var cube = [
      -1,-1, 1, 1,-1, 1, 1, 1, 1, -1,-1, 1, 1, 1, 1, -1, 1, 1,
       1,-1,-1, -1,-1,-1, -1, 1,-1, 1,-1,-1, -1, 1,-1, 1, 1,-1,
      -1,-1,-1, -1,-1, 1, -1, 1, 1, -1,-1,-1, -1, 1, 1, -1, 1,-1,
       1,-1, 1, 1,-1,-1, 1, 1,-1, 1,-1, 1, 1, 1,-1, 1, 1, 1,
      -1, 1, 1, 1, 1, 1, 1, 1,-1, -1, 1, 1, 1, 1,-1, -1, 1,-1,
      -1,-1,-1, 1,-1,-1, 1,-1, 1, -1,-1,-1, 1,-1, 1, -1,-1, 1
    ];

    function resize() {
      var rect = canvas.getBoundingClientRect();
      var dpr = Math.min(window.devicePixelRatio || 1, 2);
      var w = Math.max(320, Math.floor(rect.width * dpr));
      var h = Math.max(240, Math.floor(rect.height * dpr));
      if (canvas.width !== w || canvas.height !== h) {
        canvas.width = w;
        canvas.height = h;
      }
      gl.viewport(0, 0, canvas.width, canvas.height);
      return canvas.width / canvas.height;
    }

    function perspective(aspect) {
      var fov = Math.PI / 3.2;
      var near = 0.1;
      var far = 120;
      var f = 1 / Math.tan(fov / 2);
      var nf = 1 / (near - far);
      return [
        f / aspect, 0, 0, 0,
        0, f, 0, 0,
        0, 0, (far + near) * nf, -1,
        0, 0, (2 * far * near) * nf, 0
      ];
    }

    function multiply(a, b) {
      var a00 = a[0], a01 = a[1], a02 = a[2], a03 = a[3];
      var a10 = a[4], a11 = a[5], a12 = a[6], a13 = a[7];
      var a20 = a[8], a21 = a[9], a22 = a[10], a23 = a[11];
      var a30 = a[12], a31 = a[13], a32 = a[14], a33 = a[15];
      var b00 = b[0], b01 = b[1], b02 = b[2], b03 = b[3];
      var b10 = b[4], b11 = b[5], b12 = b[6], b13 = b[7];
      var b20 = b[8], b21 = b[9], b22 = b[10], b23 = b[11];
      var b30 = b[12], b31 = b[13], b32 = b[14], b33 = b[15];
      return [
        b00 * a00 + b01 * a10 + b02 * a20 + b03 * a30,
        b00 * a01 + b01 * a11 + b02 * a21 + b03 * a31,
        b00 * a02 + b01 * a12 + b02 * a22 + b03 * a32,
        b00 * a03 + b01 * a13 + b02 * a23 + b03 * a33,
        b10 * a00 + b11 * a10 + b12 * a20 + b13 * a30,
        b10 * a01 + b11 * a11 + b12 * a21 + b13 * a31,
        b10 * a02 + b11 * a12 + b12 * a22 + b13 * a32,
        b10 * a03 + b11 * a13 + b12 * a23 + b13 * a33,
        b20 * a00 + b21 * a10 + b22 * a20 + b23 * a30,
        b20 * a01 + b21 * a11 + b22 * a21 + b23 * a31,
        b20 * a02 + b21 * a12 + b22 * a22 + b23 * a32,
        b20 * a03 + b21 * a13 + b22 * a23 + b23 * a33,
        b30 * a00 + b31 * a10 + b32 * a20 + b33 * a30,
        b30 * a01 + b31 * a11 + b32 * a21 + b33 * a31,
        b30 * a02 + b31 * a12 + b32 * a22 + b33 * a32,
        b30 * a03 + b31 * a13 + b32 * a23 + b33 * a33
      ];
    }

    function translate(m, x, y, z) {
      return multiply(m, [1,0,0,0, 0,1,0,0, 0,0,1,0, x,y,z,1]);
    }

    function scale(m, x, y, z) {
      return multiply(m, [x,0,0,0, 0,y,0,0, 0,0,z,0, 0,0,0,1]);
    }

    function rotateY(m, a) {
      var c = Math.cos(a);
      var s = Math.sin(a);
      return multiply(m, [c,0,-s,0, 0,1,0,0, s,0,c,0, 0,0,0,1]);
    }

    function rotateX(m, a) {
      var c = Math.cos(a);
      var s = Math.sin(a);
      return multiply(m, [1,0,0,0, 0,c,s,0, 0,-s,c,0, 0,0,0,1]);
    }

    function clear(color) {
      var aspect = resize();
      gl.clearColor(color[0], color[1], color[2], 1);
      gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
      return perspective(aspect);
    }

    function drawCube(matrix, x, y, z, sx, sy, sz, color, rx, ry) {
      var m = translate(matrix, x, y, z);
      if (rx) m = rotateX(m, rx);
      if (ry) m = rotateY(m, ry);
      m = scale(m, sx, sy, sz);
      var data = [];
      for (var i = 0; i < cube.length; i += 3) {
        data.push(cube[i], cube[i + 1], cube[i + 2], color[0], color[1], color[2]);
      }
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STREAM_DRAW);
      gl.uniformMatrix4fv(matrixLoc, false, new Float32Array(m));
      gl.drawArrays(gl.TRIANGLES, 0, cube.length / 3);
    }

    return {
      clear: clear,
      drawCube: drawCube,
      multiply: multiply,
      translate: translate,
      rotateX: rotateX,
      rotateY: rotateY,
      scale: scale
    };
  }

  function bindKeys() {
    var keys = {};
    window.addEventListener("keydown", function (event) {
      keys[event.key] = true;
      if (["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", " "].indexOf(event.key) >= 0) {
        event.preventDefault();
      }
    });
    window.addEventListener("keyup", function (event) { keys[event.key] = false; });
    return keys;
  }

  global.PCSWebGLArcade = {
    createRenderer: createRenderer,
    bindKeys: bindKeys
  };
}(window));
