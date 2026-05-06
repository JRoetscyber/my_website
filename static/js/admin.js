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
    var form       = document.getElementById('portfolioForm');
    var deployBtn  = document.getElementById('deployBtn');
    var labelEl    = document.getElementById('deployBtnLabel');
    var cancelBtn  = document.getElementById('cancelEditBtn');
    var widgetTitle = document.querySelector('#projectsSection .widget-title');
    if (!form || !deployBtn) return;

    function resetForm() {
      form.reset();
      document.getElementById('editProjectId').value = '';
      if (labelEl) labelEl.textContent = 'Upload Project';
      if (cancelBtn) cancelBtn.style.display = 'none';
      if (widgetTitle) widgetTitle.childNodes[widgetTitle.childNodes.length - 1].textContent = ' Upload Project';
      document.getElementById('fileDropFilename').textContent = '';
    }

    if (cancelBtn) {
      cancelBtn.addEventListener('click', function () {
        resetForm();
      });
    }

    document.querySelectorAll('.btn-edit-project').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var d = btn.dataset;
        document.getElementById('editProjectId').value    = d.id;
        document.getElementById('project_title').value   = d.title        || '';
        document.getElementById('project_tag').value     = d.tag          || '';
        document.getElementById('description').value     = d.description  || '';
        document.getElementById('youtube_url').value     = d.youtube      || '';
        document.getElementById('project_url').value     = d.url          || '';
        document.getElementById('techStackValue').value  = d.tech         || '';
        document.getElementById('performance_score').value = d.performance || '';
        document.getElementById('seo_score').value       = d.seo          || '';
        if (labelEl) labelEl.textContent = 'Save Changes';
        if (cancelBtn) cancelBtn.style.display = '';
        if (widgetTitle) widgetTitle.childNodes[widgetTitle.childNodes.length - 1].textContent = ' Edit Project';
        document.getElementById('projectsSection').scrollIntoView({ behavior: 'smooth', block: 'start' });
      });
    });

    form.addEventListener('submit', function (e) {
      e.preventDefault();

      var projectId = document.getElementById('editProjectId').value;
      var isEdit    = !!projectId;
      var url       = isEdit ? '/admin/update_project/' + projectId : '/admin/add_project';

      if (labelEl) labelEl.textContent = isEdit ? 'Saving...' : 'Uploading...';
      deployBtn.disabled = true;

      fetch(url, { method: 'POST', body: new FormData(form) })
      .then(function (r) { return r.json(); })
      .then(function (data) {
        if (data.status === 'success') {
          if (labelEl) labelEl.textContent = isEdit ? 'Saved!' : 'Uploaded!';
          deployBtn.style.background = 'var(--status-green)';
          setTimeout(function () { window.location.reload(); }, 900);
        } else {
          if (labelEl) labelEl.textContent = 'Error';
          deployBtn.style.background = 'var(--badge-new)';
          alert((isEdit ? 'Update' : 'Upload') + ' failed: ' + data.message);
          setTimeout(function () {
            if (labelEl) labelEl.textContent = isEdit ? 'Save Changes' : 'Upload Project';
            deployBtn.disabled = false;
            deployBtn.style.background = '';
          }, 2000);
        }
      })
      .catch(function (err) {
        console.error(err);
        if (labelEl) labelEl.textContent = 'Error';
        setTimeout(function () {
          if (labelEl) labelEl.textContent = isEdit ? 'Save Changes' : 'Upload Project';
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
    var addForm = document.getElementById('addLeadForm');
    var editForm = document.getElementById('editLeadForm');
    var toastStack = document.getElementById('toastStack');
    var addLeadModalEl = document.getElementById('addLeadModal');
    var editLeadModalEl = document.getElementById('editLeadModal');
    var board = document.getElementById('crmKanbanBoard');

    if (!addForm || !editForm || !window.bootstrap || !board) return;

    var addLeadModal = addLeadModalEl ? new bootstrap.Modal(addLeadModalEl) : null;
    var editLeadModal = editLeadModalEl ? new bootstrap.Modal(editLeadModalEl) : null;
    var draggedCard = null;

    function showToast(message, level) {
      if (!toastStack) {
        window.alert(message);
        return;
      }

      var tone = level === 'success' ? 'success' : level === 'warning' ? 'warning' : 'danger';
      var toastEl = document.createElement('div');
      toastEl.className = 'toast align-items-center text-bg-' + tone + ' border-0';
      toastEl.setAttribute('role', 'alert');
      toastEl.setAttribute('aria-live', 'assertive');
      toastEl.setAttribute('aria-atomic', 'true');
      toastEl.innerHTML =
        '<div class="d-flex">' +
          '<div class="toast-body">' + message + '</div>' +
          '<button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast" aria-label="Close"></button>' +
        '</div>';

      toastStack.appendChild(toastEl);
      var toast = new bootstrap.Toast(toastEl, { delay: 3200 });
      toast.show();
      toastEl.addEventListener('hidden.bs.toast', function () {
        toastEl.remove();
      });
    }

    function fetchJson(url, options) {
      return fetch(url, options).then(function (response) {
        return response.json().then(function (data) {
          if (!response.ok || (data && data.status === 'error')) {
            var message = data && data.message ? data.message : 'Request failed';
            throw new Error(message);
          }
          return data;
        });
      });
    }

    function setButtonLoading(button, text) {
      button.dataset.originalText = button.dataset.originalText || button.innerHTML;
      button.disabled = true;
      button.innerHTML = text;
    }

    function resetButton(button) {
      button.disabled = false;
      if (button.dataset.originalText) {
        button.innerHTML = button.dataset.originalText;
      }
    }

    function formatBudget(value) {
      if (value === null || value === undefined || value === '' || Number(value) === 0) return '-';
      return 'R' + Number(value).toLocaleString(undefined, { maximumFractionDigits: 0 });
    }

    function formatDate(value) {
      if (!value) return '-';
      var date = new Date(value);
      if (Number.isNaN(date.getTime())) return '-';
      return date.toLocaleDateString();
    }

    function normalisePhoneNumber(phone) {
      if (!phone) return '';
      var digits = String(phone).replace(/\D/g, '');
      if (digits.startsWith('27')) return digits;
      if (digits.startsWith('0')) return '27' + digits.slice(1);
      return '27' + digits;
    }

    function statusClass(status) {
      return {
        'New': 'status-new',
        'Contacted': 'status-contacted',
        'Negotiating': 'status-proposal',
        'Closed': 'status-closed'
      }[status] || 'status-new';
    }

    function heatMeta(score) {
      if (score >= 80) return { label: 'Hot', className: 'lead-temp-hot' };
      if (score >= 50) return { label: 'Warm', className: 'lead-temp-warm' };
      return { label: 'Cold', className: 'lead-temp-cold' };
    }

    function createBreakdownHtml(lead) {
      var breakdown = lead.breakdown || {};
      return (
        '<details class="lead-breakdown">' +
          '<summary>Score Breakdown</summary>' +
          '<div class="lead-breakdown-grid">' +
            '<div><span>Explicit</span><strong>' + (breakdown.explicit != null ? breakdown.explicit : '-') + '</strong></div>' +
            '<div><span>Implicit</span><strong>' + (breakdown.implicit != null ? breakdown.implicit : '-') + '</strong></div>' +
            '<div><span>Urgency</span><strong>' + (breakdown.urgency != null ? breakdown.urgency : '-') + '</strong></div>' +
          '</div>' +
        '</details>'
      );
    }

    function staleMarkup(lastActivityDate) {
      if (!lastActivityDate) return '';
      var last = new Date(lastActivityDate);
      if (Number.isNaN(last.getTime())) return '';
      var diffDays = Math.floor((Date.now() - last.getTime()) / 86400000);
      if (diffDays <= 14) return '';
      return '<span class="lead-urgency-flag"><span class="lead-urgency-dot"></span>Follow up</span>';
    }

    function updateColumnCounts() {
      document.querySelectorAll('.kanban-column').forEach(function (column) {
        var countEl = column.querySelector('[data-column-count]');
        if (countEl) {
          countEl.textContent = column.querySelectorAll('.lead-kanban-card').length;
        }
      });
    }

    function buildWhatsAppUrl(phone, name, projectType) {
      var number = normalisePhoneNumber(phone);
      if (!number) return '';
      var message = 'Hi ' + (name || 'there') + ', following up on your ' + (projectType || 'project') + ' project...';
      return 'https://wa.me/' + number + '?text=' + encodeURIComponent(message);
    }

    function attachCardEvents(card) {
      card.addEventListener('dragstart', function () {
        draggedCard = card;
        card.classList.add('dragging');
      });

      card.addEventListener('dragend', function () {
        card.classList.remove('dragging');
        draggedCard = null;
        board.querySelectorAll('.kanban-dropzone').forEach(function (zone) {
          zone.classList.remove('drag-over');
        });
      });

      var editBtn = card.querySelector('.btn-edit-lead');
      if (editBtn) {
        editBtn.addEventListener('click', function () {
          var id = editBtn.getAttribute('data-id');
          setButtonLoading(editBtn, 'Loading...');
          fetchJson('/admin/get_lead/' + id)
            .then(function (data) {
              document.getElementById('editLeadId').value = data.id;
              document.getElementById('editLeadName').value = data.client_name || '';
              document.getElementById('editLeadCompany').value = data.client_company || '';
              document.getElementById('editLeadProjectType').value = data.project_type || data.target_project || 'Static';
              document.getElementById('editLeadRole').value = data.contact_role || 'Agent';
              document.getElementById('editLeadBudget').value = data.budget != null ? data.budget : '';
              document.getElementById('editLeadPhone').value = data.phone_number || '';
              document.getElementById('editLeadScore').value = data.score != null ? data.score : '';
              document.getElementById('editLeadWhatsApp').value = data.whatsapp_engagement || 'Ignored';
              document.getElementById('editLeadStatus').value = data.status || 'New';
              if (editLeadModal) editLeadModal.show();
            })
            .catch(function (err) {
              console.error(err);
              showToast(err.message, 'danger');
            })
            .finally(function () {
              resetButton(editBtn);
            });
        });
      }

      var deleteBtn = card.querySelector('.btn-delete-lead');
      if (deleteBtn) {
        deleteBtn.addEventListener('click', function () {
          var id = deleteBtn.getAttribute('data-id');
          if (!confirm('Delete this lead permanently?')) return;
          setButtonLoading(deleteBtn, 'Deleting...');
          fetchJson('/admin/delete_lead/' + id, { method: 'DELETE' })
            .then(function (data) {
              card.remove();
              updateColumnCounts();
              showToast(data.message || 'Lead deleted successfully', 'success');
            })
            .catch(function (err) {
              console.error(err);
              showToast(err.message, 'danger');
              resetButton(deleteBtn);
            });
        });
      }

      var waBtn = card.querySelector('.btn-whatsapp-lead');
      if (waBtn) {
        waBtn.addEventListener('click', function () {
          var url = buildWhatsAppUrl(
            waBtn.getAttribute('data-phone'),
            waBtn.getAttribute('data-name'),
            waBtn.getAttribute('data-project')
          );
          if (!url) {
            showToast('No phone number available for this lead', 'warning');
            return;
          }
          window.open(url, '_blank', 'noopener');
        });
      }
    }

    function renderLeadCard(lead) {
      var heat = heatMeta(Number(lead.score) || 0);
      var card = document.createElement('article');
      card.className = 'lead-kanban-card';
      card.setAttribute('draggable', 'true');
      card.setAttribute('data-lead-id', lead.id);
      card.setAttribute('data-status', lead.status || 'New');
      card.innerHTML =
        '<div class="lead-card-topline">' +
          '<span class="status-badge ' + statusClass(lead.status || 'New') + '">' + (lead.status || 'New') + '</span>' +
          staleMarkup(lead.last_activity_date) +
        '</div>' +
        '<div class="lead-card-main">' +
          '<h3 class="lead-card-company">' + (lead.client_company || lead.client_name || 'Untitled Lead') + '</h3>' +
          '<div class="lead-card-contact">' + (lead.client_name || 'No contact name') + '</div>' +
          '<div class="lead-card-project">' + (lead.project_type || lead.target_project || 'General project') + '</div>' +
        '</div>' +
        '<div class="lead-card-metrics">' +
          '<span class="lead-score-pill">' + (lead.score != null ? lead.score : '-') + '</span>' +
          '<span class="lead-temp-badge ' + heat.className + '">' + heat.label + '</span>' +
          '<span class="lead-budget">' + formatBudget(lead.budget) + '</span>' +
        '</div>' +
        '<div class="lead-card-meta">' +
          '<span>' + (lead.contact_role || 'Role n/a') + '</span>' +
          '<span>' + (lead.whatsapp_engagement || 'No WA signal') + '</span>' +
        '</div>' +
        createBreakdownHtml(lead) +
        '<div class="lead-card-actions">' +
          '<button class="tbl-action-btn btn-edit-lead" data-id="' + lead.id + '" type="button">Edit</button>' +
          '<button class="tbl-action-btn btn-whatsapp-lead" data-phone="' + (lead.phone_number || '') + '" data-name="' + (lead.client_name || lead.client_company || 'there') + '" data-project="' + (lead.project_type || lead.target_project || 'project') + '" type="button">WhatsApp</button>' +
          '<button class="tbl-action-btn btn-delete-lead" data-id="' + lead.id + '" type="button">Delete</button>' +
        '</div>';
      attachCardEvents(card);
      return card;
    }

    board.querySelectorAll('.lead-kanban-card').forEach(attachCardEvents);

    board.querySelectorAll('.kanban-dropzone').forEach(function (zone) {
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
        if (!draggedCard) return;

        var newStatus = zone.closest('.kanban-column').getAttribute('data-status');
        var leadId = draggedCard.getAttribute('data-lead-id');
        var previousZone = draggedCard.parentElement;
        var previousStatus = draggedCard.getAttribute('data-status');

        if (newStatus === previousStatus) return;

        zone.insertBefore(draggedCard, zone.querySelector('.kanban-empty'));
        updateColumnCounts();

        fetchJson('/admin/update_lead/' + leadId, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ status: newStatus })
        })
        .then(function (data) {
          if (data.lead) {
            var replacement = renderLeadCard(data.lead);
            draggedCard.replaceWith(replacement);
          }
          updateColumnCounts();
          showToast(data.message || 'Lead updated successfully', 'success');
        })
        .catch(function (err) {
          console.error(err);
          previousZone.insertBefore(draggedCard, previousZone.querySelector('.kanban-empty'));
          draggedCard.setAttribute('data-status', previousStatus);
          updateColumnCounts();
          showToast(err.message, 'danger');
        });
      });
    });

    addForm.addEventListener('submit', function (e) {
      e.preventDefault();
      var submitBtn = addForm.querySelector('button[type="submit"]');
      setButtonLoading(submitBtn, 'Saving...');

      fetchJson('/admin/add_lead', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          client_name: addForm.client_name.value.trim(),
          client_company: addForm.client_company.value.trim(),
          project_type: addForm.project_type.value,
          target_project: addForm.project_type.value,
          budget: addForm.budget.value ? parseFloat(addForm.budget.value) : 0,
          contact_role: addForm.contact_role.value,
          phone_number: addForm.phone_number.value.trim(),
          whatsapp_engagement: addForm.whatsapp_engagement.value,
          status: addForm.status.value
        })
      })
      .then(function (data) {
        showToast(data.message || 'Lead added successfully', 'success');
        addForm.reset();
        if (addLeadModal) addLeadModal.hide();
        if (data.lead) {
          var targetZone = board.querySelector('.kanban-column[data-status="' + (data.lead.status || 'New') + '"] .kanban-dropzone');
          if (targetZone) {
            targetZone.insertBefore(renderLeadCard(data.lead), targetZone.querySelector('.kanban-empty'));
            updateColumnCounts();
          }
        }
        resetButton(submitBtn);
      })
      .catch(function (err) {
        console.error(err);
        showToast(err.message, 'danger');
        resetButton(submitBtn);
      });
    });

    editForm.addEventListener('submit', function (e) {
      e.preventDefault();
      var id = document.getElementById('editLeadId').value;
      var submitBtn = editForm.querySelector('button[type="submit"]');
      var parsedScore = parseInt(document.getElementById('editLeadScore').value, 10);
      var payload = {
        status: document.getElementById('editLeadStatus').value
      };

      if (!Number.isNaN(parsedScore)) {
        payload.score = parsedScore;
      }

      payload.client_name = document.getElementById('editLeadName').value.trim();
      payload.client_company = document.getElementById('editLeadCompany').value.trim();
      payload.project_type = document.getElementById('editLeadProjectType').value;
      payload.contact_role = document.getElementById('editLeadRole').value;
      payload.budget = document.getElementById('editLeadBudget').value ? parseFloat(document.getElementById('editLeadBudget').value) : 0;
      payload.phone_number = document.getElementById('editLeadPhone').value.trim();
      payload.whatsapp_engagement = document.getElementById('editLeadWhatsApp').value;

      setButtonLoading(submitBtn, 'Updating...');

      fetchJson('/admin/update_lead/' + id, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      })
      .then(function (data) {
        showToast(data.message || 'Lead updated successfully', 'success');
        if (editLeadModal) editLeadModal.hide();
        if (data.lead) {
          var existing = board.querySelector('.lead-kanban-card[data-lead-id="' + id + '"]');
          var targetZone = board.querySelector('.kanban-column[data-status="' + (data.lead.status || 'New') + '"] .kanban-dropzone');
          if (existing) existing.remove();
          if (targetZone) {
            targetZone.insertBefore(renderLeadCard(data.lead), targetZone.querySelector('.kanban-empty'));
          }
          updateColumnCounts();
        }
        resetButton(submitBtn);
      })
      .catch(function (err) {
        console.error(err);
        showToast(err.message, 'danger');
        resetButton(submitBtn);
      });
    });

    updateColumnCounts();
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

  (function () {
    var ctx = document.getElementById('pipelineFunnelChart');
    if (!ctx || !window.Chart || !window.pipelineFunnelData) return;

    var funnel = window.pipelineFunnelData;

    new Chart(ctx, {
      type: 'bar',
      data: {
        labels: funnel.labels,
        datasets: [{
          label: 'Leads',
          data: funnel.values,
          backgroundColor: [
            'rgba(0, 229, 204, 0.78)',
            'rgba(245, 160, 32, 0.78)',
            'rgba(167, 139, 250, 0.78)',
            'rgba(34, 197, 94, 0.78)'
          ],
          borderColor: [
            '#00e5cc',
            '#f5a020',
            '#a78bfa',
            '#22c55e'
          ],
          borderWidth: 1.5,
          borderRadius: 10,
          borderSkipped: false
        }]
      },
      options: {
        indexAxis: 'y',
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
          x: {
            beginAtZero: true,
            grid: { color: 'rgba(255,255,255,0.03)' },
            ticks: { color: '#555', font: { family: 'JetBrains Mono', size: 10 } }
          },
          y: {
            grid: { display: false },
            ticks: { color: '#999', font: { family: 'JetBrains Mono', size: 10 } }
          }
        }
      }
    });
  }());
}());
