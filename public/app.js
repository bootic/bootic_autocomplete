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

  var AjaxSearch = function (url, fn) {
    this._url = url;
    this._fn = fn;
  }

  AjaxSearch.prototype = {
    run: function (params) {
      var fn = this._fn;
      var query = [];
      for(var k in params) {
        query.push([k, encodeURI(params[k])].join('='))
      }

      fetch(this._url + '?' + query.join('&')).catch(handleError).then(json).then(fn);
    }
  }

  var WsSearch = function (url, fn) {
    this._url = url.replace(/^http(\w?):/, 'ws:').replace(/\/search$/, '/ws');
    this._fn = fn;
    this._connected = false;
    this._ws = new WebSocket(this._url);
    this._ws.onopen = this.onOpen.bind(this);
    this._ws.onclose = this.onClose.bind(this);
    this._ws.onmessage = this.onMessage.bind(this);
  }

  WsSearch.prototype = {
    run: function (params) {
      this._ws.send(JSON.stringify(params));
    },
    onOpen: function (evt) {
      this._connected = true;
      console.log('open', evt)
    },
    onClose: function (evt) {
      console.log('close', evt)
    },
    onMessage: function (evt) {
      var data = JSON.parse(evt.data);
      this._fn(data);
    }
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

  function start (opts) {
    var form = opts.form,
        target = opts.target,
        template = null;

    if(opts.template) {
      tim.templates('template', document.querySelector(opts.template).innerHTML);
      template = 'template';
    }

    var render = renderInto(target, template);

    var searchConstructor = ('WebSocket' in global) ? WsSearch : AjaxSearch;
    var search = new searchConstructor(form.action, render);

    function submit (evt) {
      evt.preventDefault()
      var data = formData(form);
      search.run(data);
      return false
    }

    form.addEventListener('submit', submit);
    form.addEventListener('keyup', submit);
  }

  return {
    start: start
  }

})(window, undefined, document);


