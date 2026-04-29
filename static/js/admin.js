(function () {
  'use strict';

  (function () {
    var el = document.getElementById('topbarDate');
    if (!el) return;

    function update() {
      var now = new Date();
      var days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
      var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      var h = now.getHours().toString().padStart(2, '0');
      var m = now.getMinutes().toString().padStart(2, '0');
      el.textContent =
        days[now.getDay()] + ', ' +
        now.getDate() + ' ' + months[now.getMonth()] + ' ' + now.getFullYear() +
        ' - ' + h + ':' + m;
    }

    update();
    setInterval(update, 30000);
  }());

  (function () {
    var sidebar = document.getElementById('adminSidebar');
    var overlay = document.getElementById('sidebarOverlay');
    var toggleBtn = document.getElementById('sidebarToggle');
    if (!sidebar || !overlay || !toggleBtn) return;

    function openSidebar() {
      sidebar.classList.add('open');
      overlay.classList.add('open');
      document.body.style.overflow = 'hidden';
    }

    function closeSidebar() {
      sidebar.classList.remove('open');
      overlay.classList.remove('open');
      document.body.style.overflow = '';
    }

    toggleBtn.addEventListener('click', function () {
      if (sidebar.classList.contains('open')) {
        closeSidebar();
      } else {
        openSidebar();
      }
    });

    overlay.addEventListener('click', closeSidebar);

    sidebar.querySelectorAll('.sidebar-nav-item').forEach(function (item) {
      item.addEventListener('click', function () {
        if (window.innerWidth < 992) closeSidebar();
      });
    });
  }());

  (function () {
    var navItems = document.querySelectorAll('.sidebar-nav-item[data-section]');
    var activeSectionName = document.getElementById('activeSectionName');
    var sections = [];

    navItems.forEach(function (item) {
      var id = item.getAttribute('data-section');
      var el = document.getElementById(id);
      if (el) sections.push({ item: item, el: el });
    });

    if (!sections.length) return;

    function setActive(targetItem) {
      navItems.forEach(function (n) { n.classList.remove('active'); });
      if (targetItem) {
        targetItem.classList.add('active');
        if (activeSectionName) {
          activeSectionName.textContent = targetItem.textContent.trim().split('\n')[0];
        }
      }
    }

    navItems.forEach(function (item) {
      item.addEventListener('click', function () {
        var id = item.getAttribute('data-section');
        var el = document.getElementById(id);
        if (el) {
          var headerOffset = 80;
          var elementPosition = el.getBoundingClientRect().top;
          var offsetPosition = elementPosition + window.pageYOffset - headerOffset;

          window.scrollTo({
            top: offsetPosition,
            behavior: 'smooth'
          });
          setActive(item);
        }
      });
    });

    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          var matched = sections.find(function (s) { return s.el === entry.target; });
          if (matched) setActive(matched.item);
        }
      });
    }, { rootMargin: '-10% 0px -80% 0px', threshold: 0 });

    sections.forEach(function (s) { observer.observe(s.el); });
  }());

  (function () {
    var wrapper = document.getElementById('techStackWrapper');
    var input = document.getElementById('techStackInput');
    var hiddenInput = document.getElementById('techStackValue');
    if (!wrapper || !input || !hiddenInput) return;

    var tags = [];

    function updateHiddenInput() {
      hiddenInput.value = tags.join(',');
    }

    function createTag(label) {
      var div = document.createElement('div');
      div.className = 'tag-chip';
      div.innerHTML = '<span>' + label + '</span>';
      
      var closeBtn = document.createElement('button');
      closeBtn.className = 'tag-chip-remove';
      closeBtn.innerHTML = '&times;';
      closeBtn.type = 'button';
      closeBtn.onclick = function () {
        wrapper.removeChild(div);
        tags = tags.filter(function (t) { return t !== label; });
        updateHiddenInput();
      };
      
      div.appendChild(closeBtn);
      return div;
    }

    input.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') {
        e.preventDefault();
        var val = input.value.trim();
        if (val && !tags.includes(val)) {
          tags.push(val);
          wrapper.insertBefore(createTag(val), input);
          input.value = '';
          updateHiddenInput();
        }
      } else if (e.key === 'Backspace' && !input.value && tags.length > 0) {
        var lastTag = tags.pop();
        var chips = wrapper.querySelectorAll('.tag-chip');
        if (chips.length > 0) {
          wrapper.removeChild(chips[chips.length - 1]);
          updateHiddenInput();
        }
      }
    });

    wrapper.addEventListener('click', function () {
      input.focus();
    });
  }());

  (function () {
    var cards = document.querySelectorAll('[data-countup]');
    if (!cards.length) return;

    function countUp(el, target, duration) {
      var start = performance.now();
      var isFloat = target % 1 !== 0;
      (function tick(now) {
        var p = Math.min((now - start) / duration, 1);
        var eased = 1 - Math.pow(1 - p, 4);
        el.textContent = isFloat
          ? (eased * target).toFixed(1)
          : Math.round(eased * target).toLocaleString();
        if (p < 1) requestAnimationFrame(tick);
      }(performance.now()));
    }

    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          var el = entry.target;
          var targetValue = el.getAttribute('data-countup');
          var target = parseFloat(targetValue.replace(/[^0-9.]/g, '')) || 0;
          countUp(el, target, 1500);
          observer.unobserve(el);
        }
      });
    }, { threshold: 0.1 });

    cards.forEach(function (el) { observer.observe(el); });
  }());

  (function () {
    var logEl = document.getElementById('systemLog');

    function appendLog(scriptName, status) {
      if (!logEl) return;
      var now = new Date();
      var time = now.getHours().toString().padStart(2, '0') + ':' +
                 now.getMinutes().toString().padStart(2, '0') + ':' +
                 now.getSeconds().toString().padStart(2, '0');

      var entry = document.createElement('div');
      entry.className = 'log-entry';
      entry.innerHTML =
        '<span class="log-ts">' + time + '</span>' +
        '<span class="log-event">' + scriptName + '</span>' +
        '<span class="log-status ' + (status === 'ok' ? 'ok' : 'run') + '">' +
        (status === 'ok' ? '[COMPLETE]' : '[RUNNING]') + '</span>';

      logEl.insertBefore(entry, logEl.firstChild);
      while (logEl.children.length > 20) {
        logEl.removeChild(logEl.lastChild);
      }
    }

    document.querySelectorAll('.automation-btn').forEach(function (btn) {
      btn.addEventListener('click', function () {
        if (btn.classList.contains('executing')) return;

        var label = btn.querySelector('.automation-btn-label');
        var subLabel = btn.querySelector('.automation-btn-sub');
        var origLabel = label.textContent;
        var scriptName = btn.getAttribute('data-script') || origLabel;
        var duration = parseInt(btn.getAttribute('data-duration') || '1800', 10);

        btn.classList.add('executing');
        btn.classList.remove('complete');
        label.textContent = 'Executing...';
        if (subLabel) subLabel.textContent = 'Script running';
        appendLog(scriptName, 'run');

        fetch('/admin/run_script', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ script_name: scriptName })
        })
        .then(function (response) { return response.json(); })
        .then(function () {
          setTimeout(function () {
            btn.classList.remove('executing');
            btn.classList.add('complete');
            label.textContent = origLabel;
            if (subLabel) subLabel.textContent = btn.getAttribute('data-sub') || '';
            appendLog(scriptName, 'ok');
          }, duration);
        })
        .catch(function (err) {
          console.error(err);
          btn.classList.remove('executing');
          label.textContent = 'Error';
          setTimeout(function () {
            label.textContent = origLabel;
            if (subLabel) subLabel.textContent = btn.getAttribute('data-sub') || '';
          }, 2500);
        });
      });
    });
  }());

  (function () {
    var zone = document.getElementById('fileDropZone');
    var input = document.getElementById('fileInput');
    var nameEl = document.getElementById('fileDropFilename');
    if (!zone || !input) return;

    zone.addEventListener('click', function () { input.click(); });

    input.addEventListener('change', function () {
      if (input.files && input.files[0]) {
        var file = input.files[0];
        if (nameEl) {
          nameEl.textContent = file.name + ' (' + (file.size / 1024 / 1024).toFixed(2) + ' MB)';
          nameEl.style.display = 'block';
        }
        zone.style.borderColor = 'var(--accent)';
      }
    });

    zone.addEventListener('dragover', function (e) {
      e.preventDefault();
      zone.classList.add('drag-over');
    });

    zone.addEventListener('dragleave', function () {
      zone.classList.remove('drag-over');
    });

    zone.addEventListener('drop', function (e) {
      e.preventDefault();
      zone.classList.remove('drag-over');
      var files = e.dataTransfer.files;
      if (files && files[0]) {
        input.files = files;
        if (nameEl) {
          nameEl.textContent = files[0].name;
          nameEl.style.display = 'block';
        }
        zone.style.borderColor = 'var(--accent)';
      }
    });
  }());

  (function () {
    var form = document.getElementById('portfolioForm');
    var deployBtn = document.getElementById('deployBtn');
    if (!form || !deployBtn) return;

    form.addEventListener('submit', function (e) {
      e.preventDefault();

      var origText = deployBtn.innerHTML;
      deployBtn.textContent = 'Uploading...';
      deployBtn.disabled = true;

      fetch('/admin/add_project', {
        method: 'POST',
        body: new FormData(form)
      })
      .then(function (response) { return response.json(); })
      .then(function (data) {
        if (data.status === 'success') {
          deployBtn.textContent = 'Uploaded';
          deployBtn.style.background = 'var(--status-green)';
          setTimeout(function () {
            window.location.reload();
          }, 900);
        } else {
          deployBtn.textContent = 'Error';
          deployBtn.style.background = 'var(--badge-new)';
          alert('Deployment failed: ' + data.message);
          setTimeout(function () {
            deployBtn.innerHTML = origText;
            deployBtn.disabled = false;
            deployBtn.style.background = '';
          }, 2000);
        }
      })
      .catch(function (err) {
        console.error(err);
        deployBtn.textContent = 'Error';
        setTimeout(function () {
          deployBtn.innerHTML = origText;
          deployBtn.disabled = false;
          deployBtn.style.background = '';
        }, 2000);
      });
    });
  }());

  (function () {
    document.querySelectorAll('.btn-delete-project').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = btn.getAttribute('data-id');
        if (!confirm('Are you sure you want to delete this project?')) return;

        fetch('/admin/delete_project/' + id, { method: 'POST' })
          .then(function (response) { return response.json(); })
          .then(function (data) {
            if (data.status === 'success') {
              btn.closest('tr').remove();
            } else {
              alert('Error: ' + data.message);
            }
          })
          .catch(function (err) { console.error(err); });
      });
    });
  }());

  (function () {
    var form = document.getElementById('addLeadForm');
    if (!form) return;

    form.addEventListener('submit', function (e) {
      e.preventDefault();
      var submitBtn = form.querySelector('button[type="submit"]');
      var origText = submitBtn.textContent;

      submitBtn.textContent = 'Processing...';
      submitBtn.disabled = true;

      var data = {
        client_name: form.client_name.value,
        client_company: form.client_company.value,
        phone_number: form.phone_number.value,
        project_type: form.project_type.value,
        budget: parseFloat(form.budget.value) || 0,
        contact_role: form.contact_role.value,
        status: form.status.value,
        visited_pages: ['/admin']
      };

      fetch('/api/new-lead', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      })
      .then(function (response) { return response.json(); })
      .then(function (res) {
        if (res.status === 'success') {
          submitBtn.textContent = 'Saved';
          setTimeout(function () {
            var modal = bootstrap.Modal.getInstance(document.getElementById('addLeadModal'));
            modal.hide();
            window.location.reload();
          }, 1000);
        } else {
          alert('Error: ' + res.message);
          submitBtn.textContent = origText;
          submitBtn.disabled = false;
        }
      })
      .catch(function (err) {
        console.error(err);
        submitBtn.textContent = origText;
        submitBtn.disabled = false;
      });
    });
  }());

  (function () {
    document.querySelectorAll('.btn-delete-lead').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = btn.getAttribute('data-id');
        if (!confirm('Are you sure you want to delete this lead?')) return;

        fetch('/admin/delete_lead/' + id, { method: 'POST' })
          .then(function (response) { return response.json(); })
          .then(function (data) {
            if (data.status === 'success') {
              btn.closest('tr').remove();
            } else {
              alert('Error: ' + data.message);
            }
          })
          .catch(function (err) { console.error(err); });
      });
    });
  }());

  (function () {
    var modalEl = document.getElementById('editLeadModal');
    var editForm = document.getElementById('editLeadForm');
    if (!modalEl || !editForm || !window.bootstrap) return;

    var editModal = new bootstrap.Modal(modalEl);

    document.querySelectorAll('.btn-edit-lead').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = btn.getAttribute('data-id');

        fetch('/admin/get_lead/' + id)
          .then(function (response) { return response.json(); })
          .then(function (data) {
            if (data.status === 'error') {
              alert(data.message);
              return;
            }

            document.getElementById('editLeadId').value = data.id;
            document.getElementById('editLeadName').value = data.client_name;
            document.getElementById('editLeadScore').value = data.score;
            document.getElementById('editLeadStatus').value = data.status;
            editModal.show();
          })
          .catch(function (err) { console.error(err); });
      });
    });

    editForm.addEventListener('submit', function (e) {
      e.preventDefault();
      var id = document.getElementById('editLeadId').value;
      var submitBtn = editForm.querySelector('button[type="submit"]');
      var origText = submitBtn.textContent;

      submitBtn.textContent = 'Updating...';
      submitBtn.disabled = true;

      fetch('/admin/update_lead/' + id, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          score: parseInt(document.getElementById('editLeadScore').value, 10),
          status: document.getElementById('editLeadStatus').value
        })
      })
      .then(function (response) { return response.json(); })
      .then(function (res) {
        if (res.status === 'success') {
          submitBtn.textContent = 'Updated';
          setTimeout(function () {
            editModal.hide();
            window.location.reload();
          }, 800);
        } else {
          alert('Error: ' + res.message);
          submitBtn.textContent = origText;
          submitBtn.disabled = false;
        }
      })
      .catch(function (err) {
        console.error(err);
        submitBtn.textContent = origText;
        submitBtn.disabled = false;
      });
    });
  }());

  (function () {
    var ctx = document.getElementById('analyticsChart');
    if (!ctx || !window.Chart) return;

    fetch('/admin/analytics_data')
      .then(function (response) { return response.json(); })
      .then(function (data) {
        if (data.status !== 'success') return;

        new Chart(ctx, {
          type: 'line',
          data: {
            labels: data.labels,
            datasets: [{
              label: 'Views',
              data: data.data,
              borderColor: '#00e5cc',
              backgroundColor: 'rgba(0, 229, 204, 0.1)',
              borderWidth: 3,
              tension: 0.4,
              fill: true,
              pointRadius: 4,
              pointBackgroundColor: '#00e5cc',
              pointBorderColor: '#000',
              pointBorderWidth: 2,
              pointHoverRadius: 6
            }]
          },
          options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
              legend: { display: false },
              tooltip: {
                backgroundColor: '#111',
                titleFont: { family: 'JetBrains Mono', size: 12 },
                bodyFont: { family: 'JetBrains Mono', size: 12 },
                borderColor: '#252525',
                borderWidth: 1,
                padding: 10,
                displayColors: false
              }
            },
            scales: {
              y: {
                beginAtZero: true,
                grid: { color: 'rgba(255,255,255,0.03)' },
                ticks: { color: '#555', font: { family: 'JetBrains Mono', size: 10 } }
              },
              x: {
                grid: { display: false },
                ticks: { color: '#555', font: { family: 'JetBrains Mono', size: 10 } }
              }
            }
          }
        });
      })
      .catch(function (err) {
        console.error('Error loading analytics chart:', err);
      });
  }());
}());
