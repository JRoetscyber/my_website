/* ================================================================
   admin.js — "The Black Hole" Command Center
   JO4 Dev | Admin Dashboard Scripts | v1.0
================================================================ */

(function () {
  'use strict';

  /* ── 1. TOPBAR — live clock ── */
  (function () {
    var el = document.getElementById('topbarDate');
    if (!el) return;

    function update() {
      var now = new Date();
      var days  = ['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'];
      var months= ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
      var h = now.getHours().toString().padStart(2,'0');
      var m = now.getMinutes().toString().padStart(2,'0');
      el.textContent =
        days[now.getDay()] + ', ' +
        now.getDate() + ' ' + months[now.getMonth()] + ' ' + now.getFullYear() +
        ' — ' + h + ':' + m;
    }

    update();
    setInterval(update, 30000);
  }());


  /* ── 2. MOBILE SIDEBAR TOGGLE ── */
  (function () {
    var sidebar  = document.getElementById('adminSidebar');
    var overlay  = document.getElementById('sidebarOverlay');
    var toggleBtn= document.getElementById('sidebarToggle');
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
      sidebar.classList.contains('open') ? closeSidebar() : openSidebar();
    });

    overlay.addEventListener('click', closeSidebar);

    // Close on nav item click (mobile)
    sidebar.querySelectorAll('.sidebar-nav-item').forEach(function (item) {
      item.addEventListener('click', function () {
        if (window.innerWidth < 992) closeSidebar();
      });
    });
  }());


  /* ── 3. SIDEBAR — active nav item on scroll ── */
  (function () {
    var navItems = document.querySelectorAll('.sidebar-nav-item[data-section]');
    var sections = [];

    navItems.forEach(function (item) {
      var id = item.getAttribute('data-section');
      var el = document.getElementById(id);
      if (el) sections.push({ item: item, el: el });
    });

    if (!sections.length) return;

    function setActive(targetItem) {
      navItems.forEach(function (n) { n.classList.remove('active'); });
      if (targetItem) targetItem.classList.add('active');
    }

    // Smooth scroll on click
    navItems.forEach(function (item) {
      item.addEventListener('click', function () {
        var id = item.getAttribute('data-section');
        var el = document.getElementById(id);
        if (el) {
          el.scrollIntoView({ behavior: 'smooth', block: 'start' });
          setActive(item);
        }
      });
    });

    // Intersection observer for scroll-based activation
    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          var matched = sections.find(function (s) { return s.el === entry.target; });
          if (matched) setActive(matched.item);
        }
      });
    }, { rootMargin: '-30% 0px -60% 0px', threshold: 0 });

    sections.forEach(function (s) { observer.observe(s.el); });
  }());


  /* ── 4. STAT CARDS — count-up animation ── */
  (function () {
    var cards = document.querySelectorAll('[data-countup]');
    if (!cards.length) return;

    function countUp(el, target, duration) {
      var start = performance.now();
      var isFloat = target % 1 !== 0;
      (function tick(now) {
        var p = Math.min((now - start) / duration, 1);
        // ease-out cubic
        var eased = 1 - Math.pow(1 - p, 3);
        el.textContent = isFloat
          ? (eased * target).toFixed(1)
          : Math.round(eased * target).toLocaleString();
        if (p < 1) requestAnimationFrame(tick);
      }(performance.now()));
    }

    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          var el     = entry.target;
          var target = parseFloat(el.getAttribute('data-countup'));
          countUp(el, target, 1200);
          observer.unobserve(el);
        }
      });
    }, { threshold: 0.3 });

    cards.forEach(function (el) { observer.observe(el); });
  }());


  /* ── 5. AUTOMATION BUTTONS ── */
  (function () {
    var logEl = document.getElementById('systemLog');

    function appendLog(scriptName, status) {
      if (!logEl) return;
      var now  = new Date();
      var time = now.getHours().toString().padStart(2,'0') + ':' +
                 now.getMinutes().toString().padStart(2,'0') + ':' +
                 now.getSeconds().toString().padStart(2,'0');

      var entry = document.createElement('div');
      entry.className = 'log-entry';
      entry.innerHTML =
        '<span class="log-ts">' + time + '</span>' +
        '<span class="log-event">▶ ' + scriptName + '</span>' +
        '<span class="log-status ' + (status === 'ok' ? 'ok' : 'run') + '">' +
        (status === 'ok' ? '[COMPLETE]' : '[RUNNING]') + '</span>';

      logEl.insertBefore(entry, logEl.firstChild);

      // Keep max 20 log lines
      while (logEl.children.length > 20) {
        logEl.removeChild(logEl.lastChild);
      }
    }

    document.querySelectorAll('.automation-btn').forEach(function (btn) {
      btn.addEventListener('click', function () {
        if (btn.classList.contains('executing')) return;

        var label      = btn.querySelector('.automation-btn-label');
        var subLabel   = btn.querySelector('.automation-btn-sub');
        var origLabel  = label.textContent;
        var scriptName = btn.getAttribute('data-script') || origLabel;
        var duration   = parseInt(btn.getAttribute('data-duration') || '2400', 10);

        btn.classList.add('executing');
        btn.classList.remove('complete');
        label.textContent = 'Executing...';
        if (subLabel) subLabel.textContent = 'Script running — do not close';
        appendLog(scriptName, 'run');

        fetch('/admin/run_script', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ script_name: scriptName })
        })
        .then(response => response.json())
        .then(data => {
          setTimeout(function () {
            btn.classList.remove('executing');
            btn.classList.add('complete');
            label.textContent = '✓ ' + origLabel;
            if (subLabel) subLabel.textContent = 'Completed successfully';
            appendLog(scriptName, 'ok');

            setTimeout(function () {
              btn.classList.remove('complete');
              label.textContent = origLabel;
              if (subLabel) subLabel.textContent = btn.getAttribute('data-sub') || '';
            }, 3500);
          }, duration);
        })
        .catch(err => {
          console.error(err);
          btn.classList.remove('executing');
          label.textContent = 'Error';
          setTimeout(() => { label.textContent = origLabel; }, 3000);
        });
      });
    });
  }());


  /* ── 6. TECH STACK TAG INPUT ── */
  (function () {
    var wrapper  = document.getElementById('tagInputWrapper');
    var input    = document.getElementById('tagInput');
    var hidden   = document.getElementById('techStackValue');
    if (!wrapper || !input) return;

    var tags = [];

    function renderTags() {
      wrapper.querySelectorAll('.tag-chip').forEach(function (c) { c.remove(); });

      tags.forEach(function (tag, i) {
        var chip = document.createElement('span');
        chip.className = 'tag-chip';
        chip.innerHTML = tag +
          '<button type="button" class="tag-chip-remove" data-i="' + i + '" aria-label="Remove ' + tag + '">×</button>';
        wrapper.insertBefore(chip, input);
      });

      if (hidden) hidden.value = tags.join(',');
    }

    function addTag(raw) {
      var tag = raw.trim().replace(/,/g, '');
      if (!tag || tags.includes(tag)) return;
      tags.push(tag);
      renderTags();
    }

    input.addEventListener('keydown', function (e) {
      if (e.key === 'Enter' || e.key === ',') {
        e.preventDefault();
        addTag(input.value);
        input.value = '';
      }
      if (e.key === 'Backspace' && input.value === '' && tags.length) {
        tags.pop();
        renderTags();
      }
    });

    input.addEventListener('blur', function () {
      if (input.value.trim()) {
        addTag(input.value);
        input.value = '';
      }
    });

    wrapper.addEventListener('click', function (e) {
      if (e.target.classList.contains('tag-chip-remove')) {
        var i = parseInt(e.target.getAttribute('data-i'), 10);
        tags.splice(i, 1);
        renderTags();
      }
      input.focus();
    });
  }());


  /* ── 7. FILE DROP ZONE ── */
  (function () {
    var zone    = document.getElementById('fileDropZone');
    var input   = document.getElementById('fileInput');
    var nameEl  = document.getElementById('fileDropFilename');
    if (!zone || !input) return;

    zone.addEventListener('click', function () { input.click(); });

    input.addEventListener('change', function () {
      if (input.files && input.files[0]) {
        var file = input.files[0];
        if (nameEl) {
          nameEl.textContent = '✓  ' + file.name + '  (' + (file.size / 1024 / 1024).toFixed(2) + ' MB)';
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
        input.files = files; // assign to hidden input
        if (nameEl) {
          nameEl.textContent = '✓  ' + files[0].name;
          nameEl.style.display = 'block';
        }
        zone.style.borderColor = 'var(--accent)';
      }
    });
  }());


  /* ── 8. PORTFOLIO FORM — deploy to backend ── */
  (function () {
    var form      = document.getElementById('portfolioForm');
    var deployBtn = document.getElementById('deployBtn');
    if (!form || !deployBtn) return;

    form.addEventListener('submit', function (e) {
      e.preventDefault();

      var origText = deployBtn.innerHTML;
      deployBtn.textContent = 'Deploying...';
      deployBtn.disabled = true;

      var formData = new FormData(form);

      fetch('/admin/add_project', {
        method: 'POST',
        body: formData
      })
      .then(response => response.json())
      .then(data => {
        if (data.status === 'success') {
          deployBtn.textContent = '✓ Deployed';
          deployBtn.style.background = 'var(--status-green)';
          form.reset();
          // Optional: window.location.reload();
        } else {
          deployBtn.textContent = 'Error';
          deployBtn.style.background = 'var(--badge-new)';
          alert('Deployment failed: ' + data.message);
        }
        setTimeout(function () {
          deployBtn.innerHTML = origText;
          deployBtn.disabled = false;
          deployBtn.style.background = '';
        }, 3000);
      })
      .catch(err => {
        console.error(err);
        deployBtn.textContent = 'Error';
        setTimeout(() => { deployBtn.innerHTML = origText; deployBtn.disabled = false; }, 3000);
      });
    });
  }());

  /* ── 9. ADD LEAD FORM — submit to backend ── */
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
        visited_pages: ['/admin'] // Manual entry source
      };

      fetch('/api/new-lead', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      })
      .then(response => response.json())
      .then(res => {
        if (res.status === 'success') {
          submitBtn.textContent = '✓ Scored: ' + res.automated_score;
          setTimeout(() => {
            var modal = bootstrap.Modal.getInstance(document.getElementById('addLeadModal'));
            modal.hide();
            window.location.reload();
          }, 1500);
        } else {
          alert('Error: ' + res.message);
          submitBtn.textContent = origText;
          submitBtn.disabled = false;
        }
      })
      .catch(err => {
        console.error(err);
        submitBtn.textContent = origText;
        submitBtn.disabled = false;
      });
    });
  }());

  /* ── 10. DELETE LEAD — handle backend call ── */
  (function () {
    document.querySelectorAll('.btn-delete-lead').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = btn.getAttribute('data-id');
        if (confirm('Are you sure you want to delete this lead?')) {
          fetch('/admin/delete_lead/' + id, { method: 'POST' })
            .then(response => response.json())
            .then(data => {
              if (data.status === 'success') {
                btn.closest('tr').remove();
              } else {
                alert('Error: ' + data.message);
              }
            })
            .catch(err => console.error(err));
        }
      });
    });
  }());

  /* ── 12. EDIT LEAD — fetch and populate modal ── */
  (function () {
    var editModal = new bootstrap.Modal(document.getElementById('editLeadModal'));
    var editForm  = document.getElementById('editLeadForm');
    
    document.querySelectorAll('.btn-edit-lead').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var id = btn.getAttribute('data-id');
        
        fetch('/admin/get_lead/' + id)
          .then(response => response.json())
          .then(data => {
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
          .catch(err => console.error(err));
      });
    });

    if (editForm) {
      editForm.addEventListener('submit', function (e) {
        e.preventDefault();
        var id = document.getElementById('editLeadId').value;
        var submitBtn = editForm.querySelector('button[type="submit"]');
        var origText = submitBtn.textContent;
        
        submitBtn.textContent = 'Updating...';
        submitBtn.disabled = true;

        var data = {
          score: parseInt(document.getElementById('editLeadScore').value, 10),
          status: document.getElementById('editLeadStatus').value
        };

        fetch('/admin/update_lead/' + id, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data)
        })
        .then(response => response.json())
        .then(res => {
          if (res.status === 'success') {
            submitBtn.textContent = '✓ Updated';
            setTimeout(() => {
              editModal.hide();
              window.location.reload();
            }, 1000);
          } else {
            alert('Error: ' + res.message);
            submitBtn.textContent = origText;
            submitBtn.disabled = false;
          }
        })
        .catch(err => {
          console.error(err);
          submitBtn.textContent = origText;
          submitBtn.disabled = false;
        });
      });
    }
  }());

  /* ── 13. ANALYTICS CHART — Chart.js ── */
  (function () {
    var ctx = document.getElementById('analyticsChart');
    if (!ctx) return;

    fetch('/admin/analytics_data')
      .then(response => response.json())
      .then(data => {
        if (data.status === 'success') {
          new Chart(ctx, {
            type: 'line',
            data: {
              labels: data.labels,
              datasets: [{
                label: 'Views',
                data: data.data,
                borderColor: '#60a5fa',
                backgroundColor: 'rgba(96, 165, 250, 0.1)',
                borderWidth: 2,
                tension: 0.4,
                fill: true,
                pointRadius: 3,
                pointBackgroundColor: '#60a5fa'
              }]
            },
            options: {
              responsive: true,
              maintainAspectRatio: false,
              plugins: {
                legend: { display: false }
              },
              scales: {
                y: {
                  beginAtZero: true,
                  grid: { color: 'rgba(255,255,255,0.05)' },
                  ticks: { color: '#666', font: { family: 'JetBrains Mono', size: 10 } }
                },
                x: {
                  grid: { display: false },
                  ticks: { color: '#666', font: { family: 'JetBrains Mono', size: 10 } }
                }
              }
            }
          });
        }
      })
      .catch(err => console.error('Error loading analytics chart:', err));
  }());

}());
