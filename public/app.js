var Autocomplete = (function (global, undefined, document) {
                   
  var DEFAULT_TEMPLATE = '<ul>{{_embedded.items}}<li class="item">{{title}} ${{price}}</li>{{/_embedded.items}}</ul>';

  function json (r) {
    return r.json()
  }

  function handleError (r) {
    console.log('Error', r)
    return {
      json: function () {
        return {error: r, _embedded: {items: []}}
      }
    }
  }

  var AjaxSearch = function (url, fn, opts) {
    this._logger = opts.logger;
    this._url = url;
    this._fn = fn;
    this._logger.log('ajax searcher ready');
  }

  AjaxSearch.prototype = {
    run: function (params) {
      var fn = this._fn;
      var query = [];
      for(var k in params) {
        query.push([k, encodeURI(params[k])].join('='))
      }

      fetch(this._url + '?' + query.join('&'))
        .catch(handleError)
        .then(json)
        .then(fn);
    }
  }

  var WsSearch = function (url, fn, opts) {
    this._opts = opts || {secure: false}
    var scheme = this._opts.secure ? 'wss:' : 'ws:';
    this._logger = opts.logger;
    this._url = url.replace(/^http(\w?):/, scheme).replace(/\/search$/, '/ws');
    this._fn = fn;
    this._retries = 5;
    this.connected = false;
    this.connect()
  }

  WsSearch.prototype = {
    run: function (params) {
      this._ws.send(JSON.stringify(params));
    },
    connect: function () {
      this._ws = new WebSocket(this._url);
      this._ws.onopen = this.onOpen.bind(this);
      this._ws.onclose = this.onClose.bind(this);
      this._ws.onmessage = this.onMessage.bind(this);
    },
    onOpen: function (evt) {
      this.connected = true;
      this._logger.log('WebSocket searcher ready');
    },
    onClose: function (evt) {
      this.connected = false;
      this._logger.log('WebSocket closed');
      if(this._retries == 0) {
        this._logger.log('Websocket retried too many times. Giving up');
        return
      }

      this._retries--;
      global.setTimeout(this.connect.bind(this), 5000);
    },
    onMessage: function (evt) {
      var data = JSON.parse(evt.data);
      this._fn(data);
    }
  }

  var MultiSearch = function (url, fn, opts) {
    this._ajaxSearch = new AjaxSearch(url, fn, opts);
    this._wsSearch = supportsWebsockets() ? new WsSearch(url, fn, opts) : null
  }

  MultiSearch.prototype = {
    run: function (params) {
      if(this._wsSearch && this._wsSearch.connected) {
        this._wsSearch.run(params)
      } else {
        this._ajaxSearch.run(params)
      }
    }
  }

  function supportsWebsockets () {
    return ('WebSocket' in global);
  }

  function formData (form) {
    var inputs = form.querySelectorAll('input[value]');
    inputs = Array.prototype.slice.call(inputs);

    var data = {}
    inputs.forEach(function (i) {
      var v = i.name == 'q' ? i.value + '*' : i.value;
      data[i.name] = v
    })

    return data
  }

  function renderInto(e, template) {
    var template = template ? tim.templates(template) : DEFAULT_TEMPLATE;

    return function (data) {
      var output = tim(template, data);
      e.innerHTML = output;
    }
  }

  function shouldSearch (params) {
    return params.q != '*'
  }

  var Logger = function () {

  }

  Logger.prototype = {
    log: function (msg) {
      console.log('[' + (new Date()).toString() + '] ' + msg)
    }
  }

  function empty (t) {
    t.innerHTML = ''
  }

  function start (opts) {
    var form = opts.form,
        target = opts.target,
        secure = !!opts.secure,
        logger = opts.logger || new Logger(),
        template = null;

    if(opts.template) {
      tim.templates('template', document.querySelector(opts.template).innerHTML);
      template = 'template';
    }

    var render = renderInto(target, template);

    var search = new MultiSearch(form.action, render, {logger: logger,secure: secure});

    function submit (evt) {
      evt.preventDefault()
      var data = formData(form);
      if(shouldSearch(data)) {
        search.run(data);
      } else {
        empty(target)
      }
      return false
    }

    form.addEventListener('submit', submit);
    form.addEventListener('keyup', submit);
  }

  return {
    start: start
  }

})(window, undefined, document);


