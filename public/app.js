var Autocomplete = (function (global, undefined) {
                   
  var DEFAULT_TEMPLATE = '<ul>{{_embedded.items}}<li class="item">{{title}} ${{price}}</li>{{/_embedded.items}}</ul>';

  var Search = function (url) {
    this._url = url;
  }

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

  Search.prototype = {
    run: function (params, fn) {
      var query = [];
      for(var k in params) {
        query.push([k, encodeURI(params[k])].join('='))
      }

      fetch(this._url + '?' + query.join('&')).catch(handleError).then(json).then(fn);
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
        search = new Search(form.action),
        target = opts.target,
        template = null;

    if(opts.template) {
      tim.templates('template', document.querySelector(opts.template).innerHTML);
      template = 'template';
    }

    var render = renderInto(target, template);

    function submit (evt) {
      evt.preventDefault()
      var data = formData(form);
      search.run(data, render);
      return false
    }

    form.addEventListener('submit', submit);
    form.addEventListener('keyup', submit);
  }

  return {
    start: start
  }

})(window, undefined);


