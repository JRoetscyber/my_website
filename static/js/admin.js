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

  /* ── INVOICE / QUOTE GENERATOR ─────────────────────────────── */
  (function () {
    var STORE_KEY   = 'jo4_invoices';
    var CNT_INV     = 'jo4_inv_cnt';
    var CNT_QUO     = 'jo4_quo_cnt';

    var invNumber  = document.getElementById('invNumber');
    if (!invNumber) return;

    var invDate    = document.getElementById('invDate');
    var invDueDate = document.getElementById('invDueDate');
    var invStatus  = document.getElementById('invDocStatus');
    var invCName   = document.getElementById('invClientName');
    var invCCo     = document.getElementById('invClientCompany');
    var invCEmail  = document.getElementById('invClientEmail');
    var invCPhone  = document.getElementById('invClientPhone');
    var invVatOn   = document.getElementById('invVatEnabled');
    var invVatRate = document.getElementById('invVatRate');
    var invNotes   = document.getElementById('invNotes');
    var invItems   = document.getElementById('invItemsContainer');
    var invDoc     = document.getElementById('invDoc');
    var savedTbody = document.getElementById('invSavedTbody');
    var savedBadge = document.getElementById('invSavedBadge');
    var savedMeta  = document.getElementById('invSavedMeta');
    var invPage    = document.getElementById('invPage');
    var logoUrl    = invPage ? (invPage.dataset.logoUrl || '') : '';
    var bizSettings = (function () {
      try { return JSON.parse((invPage && invPage.dataset.bizSettings) || '{}'); } catch (e) { return {}; }
    }());

    var currentType = 'Invoice';
    var editingId   = null;

    /* helpers */
    function peekNum(type) {
      var key = type === 'Invoice' ? CNT_INV : CNT_QUO;
      var n   = parseInt(localStorage.getItem(key) || '0', 10) + 1;
      return (type === 'Invoice' ? 'INV-' : 'QUO-') + String(n).padStart(3, '0');
    }

    function nextNum(type) {
      var key = type === 'Invoice' ? CNT_INV : CNT_QUO;
      var n   = parseInt(localStorage.getItem(key) || '0', 10) + 1;
      localStorage.setItem(key, n);
      return (type === 'Invoice' ? 'INV-' : 'QUO-') + String(n).padStart(3, '0');
    }

    function todayStr() {
      var d = new Date();
      return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate());
    }

    function dueStr(days) {
      var d = new Date();
      d.setDate(d.getDate() + (days || 30));
      return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate());
    }

    function pad(n) { return String(n).padStart(2, '0'); }

    function fmtDate(iso) {
      if (!iso) return '';
      var p = iso.split('-');
      if (p.length < 3) return iso;
      var mo = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
      return p[2] + ' ' + mo[parseInt(p[1], 10) - 1] + ' ' + p[0];
    }

    function fmtCur(n) {
      var v = parseFloat(n) || 0;
      return 'R ' + v.toLocaleString('en-ZA', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    }

    function esc(s) {
      return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
    }

    function getItems() {
      return Array.from(invItems.querySelectorAll('.inv-line-row')).map(function (r) {
        return {
          desc:  r.querySelector('.inv-row-desc').value.trim(),
          qty:   parseFloat(r.querySelector('.inv-row-qty').value) || 0,
          price: parseFloat(r.querySelector('.inv-row-price').value) || 0
        };
      });
    }

    function totals(items) {
      var sub  = items.reduce(function (s, i) { return s + i.qty * i.price; }, 0);
      var on   = invVatOn  ? invVatOn.checked : false;
      var rate = invVatRate ? (parseFloat(invVatRate.value) || 0) : 0;
      var vat  = on ? sub * rate / 100 : 0;
      return { sub: sub, rate: rate, vat: vat, total: sub + vat, vatOn: on };
    }

    function getSaved() {
      try { return JSON.parse(localStorage.getItem(STORE_KEY) || '[]'); }
      catch (e) { return []; }
    }

    function updateBadges() {
      var c = getSaved().length;
      if (savedBadge) savedBadge.textContent = c;
      if (savedMeta) savedMeta.textContent = c + ' document' + (c === 1 ? '' : 's');
    }

    /* render preview */
    function renderPreview() {
      if (!invDoc) return;
      var items   = getItems();
      var t       = totals(items);
      var date    = invDate   ? invDate.value   : '';
      var due     = invDueDate ? invDueDate.value : '';
      var cName   = invCName  ? invCName.value.trim()  : '';
      var cCo     = invCCo    ? invCCo.value.trim()    : '';
      var cEmail  = invCEmail ? invCEmail.value.trim() : '';
      var cPhone  = invCPhone ? invCPhone.value.trim() : '';
      var notes   = invNotes  ? invNotes.value.trim()  : '';

      var logoHtml = logoUrl
        ? '<img src="' + esc(logoUrl) + '" alt="JO4 Dev" class="inv-doc-logo-img">'
        : '<div class="inv-doc-logo-text">JO4<span>.</span>DEV</div>';

      var rows = items.length
        ? items.map(function (i) {
            return '<tr>' +
              '<td>' + esc(i.desc || '—') + '</td>' +
              '<td class="inv-col-center">' + i.qty + '</td>' +
              '<td class="inv-col-right">' + fmtCur(i.price) + '</td>' +
              '<td class="inv-col-right">' + fmtCur(i.qty * i.price) + '</td>' +
            '</tr>';
          }).join('')
        : '<tr><td colspan="4" style="text-align:center;padding:20px;color:#ccc;font-style:italic;font-size:.85rem;">No items added yet</td></tr>';

      var vatRow = t.vatOn
        ? '<div class="inv-total-row"><span>VAT (' + t.rate + '%)</span><span>' + fmtCur(t.vat) + '</span></div>'
        : '';

      var notesHtml = notes
        ? '<div class="inv-doc-notes"><div class="inv-notes-label">Notes &amp; Payment Terms</div><div class="inv-notes-body">' + esc(notes).replace(/\n/g,'<br>') + '</div></div>'
        : '';

      var bankHtml = '';
      if (bizSettings.bank_name || bizSettings.account_number) {
        var bLines = [];
        if (bizSettings.bank_name)       bLines.push('<strong>Bank:</strong> '    + esc(bizSettings.bank_name));
        if (bizSettings.account_holder)  bLines.push('<strong>Holder:</strong> '  + esc(bizSettings.account_holder));
        if (bizSettings.account_number)  bLines.push('<strong>Account:</strong> ' + esc(bizSettings.account_number));
        if (bizSettings.branch_code)     bLines.push('<strong>Branch:</strong> '  + esc(bizSettings.branch_code));
        bankHtml = '<div class="inv-doc-bank"><div class="inv-notes-label">Banking Details</div><div class="inv-notes-body">' + bLines.join(' &nbsp;·&nbsp; ') + '</div></div>';
      }

      invDoc.innerHTML =
        '<div class="inv-doc-header">' +
          '<div class="inv-doc-logo-block">' + logoHtml + '</div>' +
          '<div class="inv-doc-meta">' +
            '<div class="inv-doc-type">' + currentType + '</div>' +
            '<div class="inv-doc-number">' + esc(invNumber.value || '—') + '</div>' +
            '<div class="inv-doc-dates">' +
              (date ? '<div class="inv-doc-date-row"><span>Issue Date</span><span>' + fmtDate(date) + '</span></div>' : '') +
              (due  ? '<div class="inv-doc-date-row"><span>Due Date</span><span>'   + fmtDate(due)  + '</span></div>' : '') +
            '</div>' +
          '</div>' +
        '</div>' +
        '<div class="inv-doc-parties">' +
          '<div class="inv-party"><div class="inv-party-label">From</div>' +
            '<div class="inv-party-name">' + esc(bizSettings.biz_name || 'JO4 Dev') + '</div>' +
            (bizSettings.biz_address ? '<div class="inv-party-line">' + esc(bizSettings.biz_address).replace(/\n/g,'<br>') + '</div>' : '') +
            (bizSettings.biz_phone   ? '<div class="inv-party-line">' + esc(bizSettings.biz_phone)   + '</div>' : '') +
            (bizSettings.biz_email   ? '<div class="inv-party-line">' + esc(bizSettings.biz_email)   + '</div>' : '<div class="inv-party-line">jroetscyber@gmail.com</div>') +
            (bizSettings.vat_number  ? '<div class="inv-party-line">VAT: ' + esc(bizSettings.vat_number) + '</div>' : '') +
          '</div>' +
          '<div class="inv-party"><div class="inv-party-label">Bill To</div><div class="inv-party-name">' + esc(cName || '—') + '</div>' +
            (cCo    ? '<div class="inv-party-line">' + esc(cCo)    + '</div>' : '') +
            (cEmail ? '<div class="inv-party-line">' + esc(cEmail) + '</div>' : '') +
            (cPhone ? '<div class="inv-party-line">' + esc(cPhone) + '</div>' : '') +
          '</div>' +
        '</div>' +
        '<table class="inv-doc-table"><thead><tr>' +
          '<th>Description</th><th class="inv-col-center">Qty</th><th class="inv-col-right">Unit Price</th><th class="inv-col-right">Total</th>' +
        '</tr></thead><tbody>' + rows + '</tbody></table>' +
        '<div class="inv-doc-totals">' +
          '<div class="inv-total-row"><span>Subtotal</span><span>' + fmtCur(t.sub) + '</span></div>' +
          vatRow +
          '<div class="inv-total-row inv-total-grand"><span>Total</span><span>' + fmtCur(t.total) + '</span></div>' +
        '</div>' +
        notesHtml +
        bankHtml;
    }

    /* add item row */
    function addItemRow(item) {
      item = item || {};
      var row = document.createElement('div');
      row.className = 'inv-line-row';
      row.innerHTML =
        '<input class="admin-form-input inv-row-desc"  placeholder="Service / description" value="' + esc(item.desc || '') + '">' +
        '<input class="admin-form-input inv-row-qty"   type="number" min="0" step="0.01" placeholder="1"    value="' + (item.qty   || '') + '">' +
        '<input class="admin-form-input inv-row-price" type="number" min="0" step="0.01" placeholder="0.00" value="' + (item.price || '') + '">' +
        '<button class="inv-remove-row" type="button" aria-label="Remove">' +
          '<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>' +
        '</button>';

      row.querySelector('.inv-remove-row').addEventListener('click', function () { row.remove(); renderPreview(); });
      row.querySelectorAll('input').forEach(function (inp) { inp.addEventListener('input', renderPreview); });
      invItems.appendChild(row);
      renderPreview();
    }

    /* tab switching */
    document.querySelectorAll('.inv-tab').forEach(function (tab) {
      tab.addEventListener('click', function () {
        document.querySelectorAll('.inv-tab').forEach(function (t) { t.classList.remove('active'); });
        tab.classList.add('active');
        var target = tab.getAttribute('data-inv-tab');
        document.querySelectorAll('.inv-tab-pane').forEach(function (p) { p.style.display = 'none'; });
        var pane = document.getElementById(target);
        if (pane) pane.style.display = '';
        if (target === 'invSavedPane') renderSavedTable();
        if (target === 'invBizPane') populateBizForm();
      });
    });

    /* business settings form */
    function populateBizForm() {
      var fields = {
        bizName: bizSettings.biz_name || '',
        bizAddress: bizSettings.biz_address || '',
        bizPhone: bizSettings.biz_phone || '',
        bizEmail: bizSettings.biz_email || '',
        bizVatNumber: bizSettings.vat_number || '',
        bizBankName: bizSettings.bank_name || '',
        bizAccountHolder: bizSettings.account_holder || '',
        bizAccountNumber: bizSettings.account_number || '',
        bizBranchCode: bizSettings.branch_code || '',
        bizPaymentTerms: bizSettings.payment_terms || ''
      };
      Object.keys(fields).forEach(function (id) {
        var el = document.getElementById(id);
        if (el) el.value = fields[id];
      });
    }

    var bizSaveBtn = document.getElementById('bizSaveBtn');
    if (bizSaveBtn) {
      bizSaveBtn.addEventListener('click', function () {
        var fd = new FormData();
        fd.append('biz_name',        (document.getElementById('bizName')          || {}).value || '');
        fd.append('biz_address',     (document.getElementById('bizAddress')        || {}).value || '');
        fd.append('biz_phone',       (document.getElementById('bizPhone')          || {}).value || '');
        fd.append('biz_email',       (document.getElementById('bizEmail')          || {}).value || '');
        fd.append('vat_number',      (document.getElementById('bizVatNumber')      || {}).value || '');
        fd.append('bank_name',       (document.getElementById('bizBankName')       || {}).value || '');
        fd.append('account_holder',  (document.getElementById('bizAccountHolder')  || {}).value || '');
        fd.append('account_number',  (document.getElementById('bizAccountNumber')  || {}).value || '');
        fd.append('branch_code',     (document.getElementById('bizBranchCode')     || {}).value || '');
        fd.append('payment_terms',   (document.getElementById('bizPaymentTerms')   || {}).value || '');

        bizSaveBtn.disabled = true;
        bizSaveBtn.textContent = 'Saving…';

        fetch('/admin/save_invoice_settings', { method: 'POST', body: fd })
          .then(function (r) { return r.json(); })
          .then(function (data) {
            if (data.ok) {
              bizSettings.biz_name       = fd.get('biz_name');
              bizSettings.biz_address    = fd.get('biz_address');
              bizSettings.biz_phone      = fd.get('biz_phone');
              bizSettings.biz_email      = fd.get('biz_email');
              bizSettings.vat_number     = fd.get('vat_number');
              bizSettings.bank_name      = fd.get('bank_name');
              bizSettings.account_holder = fd.get('account_holder');
              bizSettings.account_number = fd.get('account_number');
              bizSettings.branch_code    = fd.get('branch_code');
              bizSettings.payment_terms  = fd.get('payment_terms');
              renderPreview();
              bizSaveBtn.style.background = 'var(--status-green)';
              bizSaveBtn.textContent = 'Saved!';
              setTimeout(function () {
                bizSaveBtn.style.background = '';
                bizSaveBtn.disabled = false;
                bizSaveBtn.innerHTML = 'Save Settings <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>';
              }, 2000);
            } else {
              bizSaveBtn.style.background = 'var(--status-red, #e53e3e)';
              bizSaveBtn.textContent = 'Error — try again';
              setTimeout(function () { bizSaveBtn.style.background = ''; bizSaveBtn.disabled = false; bizSaveBtn.textContent = 'Save Settings'; }, 2500);
            }
          })
          .catch(function () {
            bizSaveBtn.style.background = 'var(--status-red, #e53e3e)';
            bizSaveBtn.textContent = 'Error — try again';
            setTimeout(function () { bizSaveBtn.style.background = ''; bizSaveBtn.disabled = false; bizSaveBtn.textContent = 'Save Settings'; }, 2500);
          });
      });
    }

    /* type toggle */
    document.querySelectorAll('.inv-type-btn').forEach(function (btn) {
      btn.addEventListener('click', function () {
        document.querySelectorAll('.inv-type-btn').forEach(function (b) { b.classList.remove('active'); });
        btn.classList.add('active');
        currentType = btn.getAttribute('data-type');
        var meta = document.getElementById('invDocTypeMeta');
        if (meta) meta.textContent = currentType;
        if (!editingId) invNumber.value = peekNum(currentType);
        renderPreview();
      });
    });

    /* live update on field change */
    [invDate, invDueDate, invCName, invCCo, invCEmail, invCPhone, invNotes, invVatRate].forEach(function (el) {
      if (el) el.addEventListener('input', renderPreview);
    });
    if (invVatOn) invVatOn.addEventListener('change', function () {
      if (invVatRate) invVatRate.disabled = !invVatOn.checked;
      renderPreview();
    });

    /* add row button */
    var addRowBtn = document.getElementById('invAddRow');
    if (addRowBtn) addRowBtn.addEventListener('click', function () { addItemRow(); });

    /* save */
    document.getElementById('invSaveBtn').addEventListener('click', function () {
      var items = getItems();
      var t     = totals(items);
      var docNum;

      if (editingId) {
        docNum = invNumber.value;
      } else {
        docNum = nextNum(currentType);
        invNumber.value = docNum;
      }

      var doc = {
        id:            editingId || ('inv_' + Date.now()),
        type:          currentType,
        number:        docNum,
        date:          invDate    ? invDate.value    : '',
        dueDate:       invDueDate ? invDueDate.value : '',
        status:        invStatus  ? invStatus.value  : 'Draft',
        clientName:    invCName   ? invCName.value.trim()   : '',
        clientCompany: invCCo     ? invCCo.value.trim()     : '',
        clientEmail:   invCEmail  ? invCEmail.value.trim()  : '',
        clientPhone:   invCPhone  ? invCPhone.value.trim()  : '',
        items:         items,
        vatEnabled:    t.vatOn,
        vatRate:       t.rate,
        notes:         invNotes   ? invNotes.value.trim()   : '',
        subtotal:      t.sub,
        vatAmount:     t.vat,
        total:         t.total,
        savedAt:       todayStr()
      };

      var saved = getSaved();
      if (editingId) {
        var idx = saved.findIndex(function (d) { return d.id === editingId; });
        if (idx >= 0) saved[idx] = doc; else saved.push(doc);
      } else {
        saved.push(doc);
        editingId = doc.id;
      }

      localStorage.setItem(STORE_KEY, JSON.stringify(saved));
      updateBadges();

      var btn = document.getElementById('invSaveBtn');
      var orig = btn.innerHTML;
      btn.style.background = 'var(--status-green)';
      btn.style.color = '#fff';
      btn.textContent = 'Saved!';
      setTimeout(function () { btn.style.background = ''; btn.style.color = ''; btn.innerHTML = orig; }, 1800);
    });

    /* print / PDF */
    document.getElementById('invPrintBtn').addEventListener('click', function () {
      renderPreview();
      var docEl = document.getElementById('invDoc');
      if (!docEl) return;

      var w = window.open('', '_blank', 'width=860,height=720');
      if (!w) { alert('Allow popups to print.'); return; }

      var css = [
        'body{margin:0;padding:40px;background:#fff;font-family:Inter,system-ui,sans-serif;-webkit-font-smoothing:antialiased;}',
        '.inv-doc{max-width:760px;margin:0 auto;}',
        '.inv-doc-header{display:flex;align-items:flex-start;justify-content:space-between;margin-bottom:32px;padding-bottom:24px;border-bottom:2px solid #eeece8;}',
        '.inv-doc-logo-img{max-height:50px;max-width:150px;object-fit:contain;}',
        '.inv-doc-logo-text{font-size:1.7rem;font-weight:800;letter-spacing:-0.04em;color:#111;}',
        '.inv-doc-logo-text span{color:#0d9488;}',
        '.inv-doc-meta{text-align:right;}',
        '.inv-doc-type{font-size:1.8rem;font-weight:800;text-transform:uppercase;color:#111;line-height:1;}',
        '.inv-doc-number{font-family:monospace;font-size:.8rem;color:#0d9488;letter-spacing:.06em;margin-top:6px;}',
        '.inv-doc-dates{margin-top:12px;}',
        '.inv-doc-date-row{display:flex;justify-content:flex-end;gap:14px;font-size:.82rem;color:#444;margin-top:4px;}',
        '.inv-doc-date-row span:first-child{font-family:monospace;font-size:.66rem;text-transform:uppercase;letter-spacing:.08em;color:#bbb;min-width:76px;text-align:right;}',
        '.inv-doc-parties{display:grid;grid-template-columns:1fr 1fr;gap:32px;margin-bottom:28px;padding-bottom:24px;border-bottom:1px solid #eeece8;}',
        '.inv-party-label{font-family:monospace;font-size:.6rem;text-transform:uppercase;letter-spacing:.14em;color:#bbb;margin-bottom:8px;}',
        '.inv-party-name{font-size:1rem;font-weight:700;color:#111;margin-bottom:4px;}',
        '.inv-party-line{font-size:.82rem;color:#555;line-height:1.5;}',
        '.inv-doc-table{width:100%;border-collapse:collapse;margin-bottom:20px;}',
        '.inv-doc-table thead tr{background:#f7f6f3;}',
        '.inv-doc-table thead th{font-family:monospace;font-size:.6rem;text-transform:uppercase;letter-spacing:.1em;color:#999;padding:10px 14px;text-align:left;border-bottom:1px solid #e8e6e2;}',
        '.inv-doc-table tbody tr:nth-child(even){background:#faf9f7;}',
        '.inv-doc-table td{padding:10px 14px;font-size:.88rem;color:#333;border-bottom:1px solid #f0eee9;}',
        '.inv-col-center{text-align:center!important;}',
        '.inv-col-right{text-align:right!important;}',
        '.inv-doc-totals{display:flex;flex-direction:column;align-items:flex-end;gap:6px;padding:14px 0 4px;border-top:1px solid #e8e6e2;}',
        '.inv-total-row{display:flex;gap:44px;font-size:.88rem;color:#555;}',
        '.inv-total-row span:first-child{font-family:monospace;font-size:.68rem;text-transform:uppercase;letter-spacing:.08em;color:#bbb;min-width:90px;text-align:right;}',
        '.inv-total-row span:last-child{min-width:110px;text-align:right;}',
        '.inv-total-grand{margin-top:8px;padding-top:10px;border-top:2px solid #111;}',
        '.inv-total-grand span:first-child{font-size:.78rem;font-weight:700;color:#111!important;}',
        '.inv-total-grand span:last-child{font-size:1.15rem;font-weight:800;color:#0d9488!important;}',
        '.inv-doc-notes{margin-top:20px;padding:14px 18px;background:#f7f6f3;border-radius:6px;border-left:3px solid #0d9488;}',
        '.inv-doc-bank{margin-top:12px;padding:14px 18px;background:#f7f6f3;border-radius:6px;border-left:3px solid #ccc;}',
        '.inv-notes-label{font-family:monospace;font-size:.6rem;text-transform:uppercase;letter-spacing:.12em;color:#bbb;margin-bottom:8px;}',
        '.inv-notes-body{font-size:.82rem;color:#555;line-height:1.7;}'
      ].join('');

      w.document.write(
        '<!DOCTYPE html><html><head><meta charset="UTF-8">' +
        '<title>' + esc(invNumber.value || 'Document') + '</title>' +
        '<style>' + css + '</style></head>' +
        '<body><div class="inv-doc">' + docEl.innerHTML + '</div></body></html>'
      );
      w.document.close();
      w.focus();
      setTimeout(function () { w.print(); }, 500);
    });

    /* reset / new */
    document.getElementById('invResetBtn').addEventListener('click', function () {
      if (!confirm('Start a new document? Unsaved changes will be lost.')) return;
      resetForm();
    });

    function resetForm() {
      editingId   = null;
      currentType = 'Invoice';
      document.querySelectorAll('.inv-type-btn').forEach(function (b) {
        b.classList.toggle('active', b.getAttribute('data-type') === 'Invoice');
      });
      var meta = document.getElementById('invDocTypeMeta');
      if (meta) meta.textContent = 'Invoice';
      invNumber.value = peekNum('Invoice');
      if (invDate)    invDate.value    = todayStr();
      if (invDueDate) invDueDate.value = dueStr(30);
      if (invStatus)  invStatus.value  = 'Draft';
      if (invCName)   invCName.value   = '';
      if (invCCo)     invCCo.value     = '';
      if (invCEmail)  invCEmail.value  = '';
      if (invCPhone)  invCPhone.value  = '';
      if (invNotes)   invNotes.value   = '';
      if (invVatOn)   invVatOn.checked = true;
      if (invVatRate) { invVatRate.value = '15'; invVatRate.disabled = false; }
      invItems.innerHTML = '';
      addItemRow();
      renderPreview();
    }

    /* render saved table */
    function statusClass(s) {
      return { Draft:'inv-status-draft', Sent:'inv-status-sent', Paid:'inv-status-paid', Overdue:'inv-status-overdue' }[s] || 'inv-status-draft';
    }

    function renderSavedTable() {
      if (!savedTbody) return;
      var saved = getSaved().slice().reverse();
      updateBadges();

      if (!saved.length) {
        savedTbody.innerHTML = '<tr><td colspan="7" class="p-4 text-center" style="font-family:var(--font-mono);font-size:.8rem;color:var(--color-text-muted);">No documents saved yet.</td></tr>';
        return;
      }

      savedTbody.innerHTML = saved.map(function (d) {
        return '<tr>' +
          '<td><span style="font-family:var(--font-mono);font-size:.75rem;color:var(--accent);">' + esc(d.number) + '</span></td>' +
          '<td><div class="client-name">' + esc(d.clientName || '—') + '</div><div class="client-company">' + esc(d.clientCompany || '') + '</div></td>' +
          '<td><span style="font-family:var(--font-mono);font-size:.72rem;color:var(--color-text-secondary);">' + esc(d.type) + '</span></td>' +
          '<td><span class="text-muted" style="font-family:var(--font-mono);font-size:.72rem;">' + esc(fmtDate(d.date) || '—') + '</span></td>' +
          '<td><span style="font-family:var(--font-mono);font-size:.82rem;font-weight:600;color:var(--color-text-primary);">' + fmtCur(d.total) + '</span></td>' +
          '<td><span class="inv-status-badge ' + statusClass(d.status) + '">' + esc(d.status) + '</span></td>' +
          '<td style="white-space:nowrap;">' +
            '<button class="tbl-action-btn inv-open-btn" data-doc-id="' + esc(d.id) + '" type="button">Open</button>&nbsp;' +
            '<button class="tbl-action-btn inv-del-btn" data-doc-id="' + esc(d.id) + '" type="button">Delete</button>' +
          '</td>' +
        '</tr>';
      }).join('');

      savedTbody.querySelectorAll('.inv-open-btn').forEach(function (btn) {
        btn.addEventListener('click', function () {
          var id  = btn.getAttribute('data-doc-id');
          var doc = getSaved().find(function (d) { return d.id === id; });
          if (!doc) return;
          loadDoc(doc);
          document.querySelector('.inv-tab[data-inv-tab="invBuilderPane"]').click();
        });
      });

      savedTbody.querySelectorAll('.inv-del-btn').forEach(function (btn) {
        btn.addEventListener('click', function () {
          if (!confirm('Delete this document permanently?')) return;
          var id = btn.getAttribute('data-doc-id');
          localStorage.setItem(STORE_KEY, JSON.stringify(getSaved().filter(function (d) { return d.id !== id; })));
          renderSavedTable();
        });
      });
    }

    function loadDoc(doc) {
      editingId   = doc.id;
      currentType = doc.type || 'Invoice';
      document.querySelectorAll('.inv-type-btn').forEach(function (b) {
        b.classList.toggle('active', b.getAttribute('data-type') === currentType);
      });
      var meta = document.getElementById('invDocTypeMeta');
      if (meta) meta.textContent = currentType;
      invNumber.value = doc.number || '';
      if (invDate)    invDate.value    = doc.date    || '';
      if (invDueDate) invDueDate.value = doc.dueDate || '';
      if (invStatus)  invStatus.value  = doc.status  || 'Draft';
      if (invCName)   invCName.value   = doc.clientName    || '';
      if (invCCo)     invCCo.value     = doc.clientCompany || '';
      if (invCEmail)  invCEmail.value  = doc.clientEmail   || '';
      if (invCPhone)  invCPhone.value  = doc.clientPhone   || '';
      if (invVatOn)   invVatOn.checked = doc.vatEnabled !== false;
      if (invVatRate) invVatRate.value = doc.vatRate || 15;
      if (invNotes)   invNotes.value   = doc.notes   || '';
      invItems.innerHTML = '';
      (doc.items || []).forEach(addItemRow);
      if (!doc.items || !doc.items.length) addItemRow();
      renderPreview();
    }

    /* init */
    resetForm();
    updateBadges();
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
