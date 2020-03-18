$(function() {
  /**
   * Append containers to table.
   *
   * @param {array} containers
   */
  function appendContainers(containers) {
    var $sites = $('.sites table tbody');
    var keys = ['Id', 'Image', 'State', 'Status'];
    var keep = [];
    var restartLink = '<a href="#" class="restart text-blue-600 hover:text-blue-900 focus:outline-none focus:underline">Restart</a>';
    var deleteLink = '<a href="#" class="delete text-red-600 hover:text-red-900 focus:outline-none focus:underline">Delete</a>';

    if (!containers) {
        containers = [];
    }

    for (var i = 0, l = containers.length; i < l; i++) {
      var container = containers[i];
      var url = container.Names[0].substr(1);

      var $tr = $sites.find('tr[data-container-id="' + container.Id + '"]');

      if ($tr.size()) {
        keep.push(container.Id);

        for (var key in container) {
          if (keys.indexOf(key) === -1) {
            continue;
          }

          if (key == 'Id') {
            container[key] = container[key].substr(0, 12);
          }

          $tr.find('td.container-' + key.toLowerCase()).text(container[key]);
        }
      } else {
        var html = [
          '<tr data-container-id="' +
            container.Id +
            '" class="' +
            (i % 2 !== 0 ? '' : '') +
            '">',
          '<td class="container-url px-6 py-4 whitespace-no-wrap border-b border-gray-200"><a class="text-blue-600 hover:text-blue-900 focus:outline-none focus:underline" href="//' +
            url +
            '" target="_blank">' +
            url +
            '</a></td>'
        ];

        keep.push(container.Id);

        for (var key in container) {
          if (keys.indexOf(key) === -1) {
            continue;
          }

          if (key == 'Id') {
            container[key] = container[key].substr(0, 12);
          }

          html.push(
            '<td class="container-' +
              key.toLowerCase() +
              ' border px-4 py-2"><span class="inline-flex items-center px-3 py-1 rounded text-sm font-medium leading-4 bg-gray-100 text-gray-800">' +
              container[key] +
              '</span></td>'
          );
        }

        var links = [];

        if (typeof farmer !== 'undefined') {
          var values = {
            '{id}': /(\w+\-\d+)/.exec(url),
            '{num}': /(\d+)/.exec(url),
            '{url}': url
          };

          for (var key in farmer.links) {
            var link = farmer.links[key];
            var add = true;

            for (var v in values) {
              var val = values[v];

              if (v === '{id}') {
                if (val && val.length) {
                  link = link.replace(v, val[0].toUpperCase());
                } else {
                  add = false;
                }
              } else {
                link = link.replace(v, val);
              }
            }

            if (add) {
              links.push(
                '<a class="btn" href="' +
                  link +
                  '" target="_blank">' +
                  key +
                  '</a>'
              );
            }
          }
        }

        html.push(
          '<td class="container-actions px-6 py-4 whitespace-no-wrap border-b border-gray-200 text-sm leading-5 font-medium"><div class=" flex justify-around">' + restartLink + deleteLink +
            links.join('') +
            '</div></td>'
        );
        html.push('</tr>');

        $sites.append(html.join(''));
      }
    }

    if (keep.length > 1 || !window.all) {
      $sites.find('tr').each(function() {
        var $this = $(this);

        if (keep.indexOf($this.data('container-id')) === -1) {
          $this.remove();
        }
      });
    }

    $('.loader').hide();
  }

  /**
   * Update containers.
   */
  function updateContainers() {
    window.all = typeof window.all === 'undefined' ? false : window.all;

    $('.loader').show();

    $.getJSON('/api/containers?all=' + window.all, appendContainers);
  }
  updateContainers();
  setInterval(updateContainers, 300000);

  // Delete a container.
  $(document.body).on('click', '.container-actions .delete', function(e) {
    e.preventDefault();

    var $this = $(this);
    var domain = $this
      .closest('tr')
      .find('.container-url')
      .text();
    var result = prompt('Type "delete" to confirm delete of container');

    if (result !== 'delete') {
      return;
    }

    $.getJSON('/api/containers?action=delete&domain=' + domain, function(res) {
      $('.loader').hide();

      if (res.success) {
        $this.closest('tr').remove();
      }
    });
  });

  // Restart a container.
  $(document.body).on('click', '.container-actions .restart', function(e) {
    e.preventDefault();

    var $this = $(this);
    var domain = $this
      .closest('tr')
      .find('.container-url')
      .text();

    $('.loader').show();

    $.getJSON(
      '/api/containers?action=restart&domain=' + domain,
      appendContainers
    );
  });

  // Show all/less.
  $('.show-all').on('click', function(e) {
    e.preventDefault();

    var $this = $(this);
    window.all = $this.text() === 'Show all';
    updateContainers();

    var text = $this.attr('data-text');
    $this.attr('data-text', $this.text()).text(text);
  });
});
