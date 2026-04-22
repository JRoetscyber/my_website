/* ================================================================
   projects.js — Evidence Room page scripts
   CRM Lead Scoring Engine v2 + Code block copy button
   JO4 Dev | SA-accurate, keyword-aware | 5 categories × 20 = 100
================================================================ */

/* ----------------------------------------------------------------
   1. CRM LEAD SCORING ENGINE
---------------------------------------------------------------- */
(function () {

  /* ── CATEGORY 1: Source Quality (0–20) ── */
  const SRC_SCORE = {
    referral:   20,
    whatsapp:   17,
    property24: 13,
    website:    10,
    facebook:   7,
    cold:       3,
  };

  /* ── CATEGORY 2: Engagement Velocity (0–20) ── */
  const ENG_SCORE = {
    multi_deep:      20,
    multi:           15,
    single_specific: 11,
    single_general:  6,
    none:            2,
  };

  /* ── CATEGORY 3: Purchase Readiness (0–20 base + keyword boost, cap 20) ── */
  const BUDGET_BASE = {
    cash:     20,  // pre-approved / cash buyer — maximum readiness signal
    over10m:  17,  // high value, longer cycle
    '5to10m': 18,  // SA sweet spot
    '3to5m':  15,
    '1to3m':  10,
    under1m:  5,
  };

  /* ── TIMING modifier on final score (not a category) ── */
  const TIMING_MOD = {
    biz_answered: 1.0,   // no penalty
    biz_missed:   0.92,  // -8%: voicemail = they may move on
    after_hours:  0.85,  // needs first-thing-morning contact
    weekend:      0.80,  // highest re-score risk if not followed up Monday
  };

  /* ----------------------------------------------------------------
     KEYWORD INTELLIGENCE TABLES
     Each entry: { terms[], label, cat, pts, advisory? }
     cat = 'readiness' | 'intent' | 'resilience'
  ---------------------------------------------------------------- */
  const HOT_KEYWORDS = [

    // ── Purchase Readiness boosters ──
    {
      terms: ['pre-approved', 'pre approved', 'preapproved', 'approved bond', 'bond approved'],
      label: 'Bond pre-approved', cat: 'readiness', pts: 10,
      advisory: 'Bond pre-approved — confirm amount in opening call, have OTP template ready',
    },
    {
      terms: ['pre-qualified', 'pre qualified', 'prequalified'],
      label: 'Pre-qualified buyer', cat: 'readiness', pts: 9,
      advisory: 'Pre-qualified — skip educational content, go straight to listing match',
    },
    {
      terms: ['cash buyer', 'paying cash', 'cash purchase', 'cash deal', 'cash offer'],
      label: 'Cash buyer', cat: 'readiness', pts: 10,
      advisory: 'Cash buyer — zero bond risk. Prioritise above all other leads immediately',
    },
    {
      terms: [
        'my wife and i', 'my husband and i', 'my partner and i',
        'my fiancé', 'my fiance', 'my fiancee',
        'we are looking', 'we want to', "we'd like",
        'my wife', 'my husband', 'my partner is', 'my boss', 'with my partner',
      ],
      label: 'Stakeholder involved', cat: 'readiness', pts: 6,
      advisory: 'Joint decision-maker detected — offer two viewing time slots to accommodate both',
    },
    {
      terms: ['deposit ready', 'deposit available', 'have the deposit', 'deposit in hand'],
      label: 'Deposit confirmed', cat: 'readiness', pts: 8,
      advisory: 'Deposit confirmed — move to closing sequence, not nurture sequence',
    },

    // ── Intent boosters ──
    {
      terms: [
        'arrange a viewing', 'book a viewing', 'schedule a viewing',
        'when can we view', 'when can i view', 'when can we come',
        'can we come see', 'want to view', 'like to view', 'viewing this weekend',
      ],
      label: 'Viewing request', cat: 'intent', pts: 12,
      advisory: 'Direct viewing request — secure the date BEFORE ending the call',
    },
    {
      terms: ['make an offer', 'put in an offer', 'submitting an offer', 'ready to offer', 'submit offer'],
      label: 'Offer intent', cat: 'intent', pts: 14,
      advisory: 'Offer intent detected — have the OTP and deed of sale documents on standby',
    },
    {
      terms: [' otp ', 'offer to purchase', 'sign the otp', 'sign contract', 'sign today'],
      label: 'OTP / contract', cat: 'intent', pts: 15,
      advisory: 'OTP language used — escalate to principal agent, do not leave in standard queue',
    },
    {
      terms: [
        'moving in', 'occupation date', 'move in date', 'when can i move',
        'transfer date', 'keys', 'when do i get the keys',
      ],
      label: 'Occupation urgency', cat: 'intent', pts: 11,
      advisory: 'Asking about occupation — confirm transfer timeline and sectional title register status',
    },
    {
      terms: ['available', 'still available', 'is it available', 'is the property available'],
      label: 'Availability check', cat: 'intent', pts: 5,
      advisory: null,
    },

    // ── SA Resilience Index ──
    {
      terms: [
        'solar', 'inverter', 'solar inverter', 'solar backup', 'solar panels',
        'backup power', 'load shedding', 'loadshedding', 'generator', 'ups ', 'edge certified',
      ],
      label: 'Solar / power backup', cat: 'resilience', pts: 9,
      advisory: 'Load-shedding concern — filter inventory for Edge-certified and solar-equipped units first',
    },
    {
      terms: [
        'fibre', 'fiber', 'fibre line', 'fibre ready', 'internet connection',
        'connectivity', 'work from home', 'wfh', 'openserve', 'vuma',
      ],
      label: 'Fibre connectivity', cat: 'resilience', pts: 7,
      advisory: 'Fibre requirement noted — confirm FTTH availability in each listing before presenting',
    },
    {
      terms: [
        'security', 'armed response', 'cctv', 'electric fence', 'safe area',
        'guarded estate', 'security estate', 'boomed', 'access control', 'safe neighborhood',
      ],
      label: 'Security requirements', cat: 'resilience', pts: 7,
      advisory: 'Security is a dealbreaker for this lead — only present estates with active armed response',
    },
    {
      terms: ['borehole', 'water storage', 'jojo tank', 'water backup', 'water shortage'],
      label: 'Water resilience', cat: 'resilience', pts: 6,
      advisory: 'Water resilience flagged — highlight borehole / JoJo tank installations in your pitch',
    },
  ];

  const COLD_KEYWORDS = [
    {
      terms: ['just looking', 'just browsing', 'just checking', 'just researching', 'just curious'],
      label: 'Just browsing',
    },
    {
      terms: ['maybe next year', 'in the future', 'eventually', 'someday', 'not sure yet', 'no rush', 'when the time is right'],
      label: 'No timeline',
    },
    {
      terms: ['how much does', 'rough price', 'average price', 'general price range', 'ballpark', 'what does it cost'],
      label: 'Generic price ask',
    },
    {
      terms: ['what do you do', 'how does it work', 'can you explain', 'tell me more about'],
      label: 'Awareness-stage',
    },
    {
      terms: ['is this still available', 'are you still selling', 'still for sale'],
      label: 'Generic availability check',
    },
  ];

  /* ── PRIORITY TIERS ── */
  const PRIORITY = [
    {
      min: 78, label: 'HOT LEAD', color: '#00e5cc', textColor: '#000',
      channel: 'WhatsApp FIRST — then immediate call',
      sla:     '≤ 5 minutes — every delay costs conversion probability',
      agent:   'Senior Specialist Agent',
    },
    {
      min: 52, label: 'WARM LEAD', color: '#f5a020', textColor: '#000',
      channel: 'WhatsApp intro, follow-up call within 1 hour',
      sla:     '≤ 1 hour — nurture with listing-specific content',
      agent:   'Residential Agent',
    },
    {
      min: 25, label: 'COLD LEAD', color: '#4a8fd4', textColor: '#fff',
      channel: 'Day-1 automated email sequence',
      sla:     'End of business day — agent check-in at 48h',
      agent:   'Junior Agent / Automated sequence',
    },
    {
      min: 0, label: 'LOW VALUE', color: '#444444', textColor: '#888',
      channel: 'Monthly newsletter only',
      sla:     'Automated only — re-score in 30 days',
      agent:   'No agent assigned',
    },
  ];

  /* ── DOM refs ── */
  const form           = document.getElementById('scorerForm');
  const idle           = document.getElementById('scorerIdle');
  const processing     = document.getElementById('scorerProcessing');
  const procLabel      = document.getElementById('processingLabel');
  const resultPanel    = document.getElementById('scorerResult');
  const arc            = document.getElementById('scoreArc');
  const scoreNum       = document.getElementById('scoreNumber');
  const badge          = document.getElementById('priorityBadge');
  const routingLine    = document.getElementById('routingLine');
  const confidenceLine = document.getElementById('confidenceLine');
  const signalsWrap    = document.getElementById('detectedSignalsWrap');
  const signalsRow     = document.getElementById('detectedSignals');
  const rChannel       = document.getElementById('routingChannel');
  const rAgent         = document.getElementById('routingAgent');
  const rSLA           = document.getElementById('routingSLA');
  const advisoriesEl   = document.getElementById('routingAdvisories');
  const btnLabel       = document.getElementById('scoreBtnLabel');

  if (!form) return; // guard: only run on projects page

  const CIRC = 2 * Math.PI * 54; // r=54 → 339.292

  /* ── Helpers ── */
  function cap(n, max) { return Math.min(n, max); }

  function detect(text, terms) {
    const t = text.toLowerCase();
    return terms.some(function (term) { return t.includes(term.toLowerCase()); });
  }

  function animateRing(score, color) {
    arc.style.transition = 'none';
    arc.style.strokeDashoffset = CIRC;
    arc.getBoundingClientRect(); // force reflow
    arc.style.transition = 'stroke-dashoffset 1.1s cubic-bezier(.4,0,.2,1), stroke 0.4s ease';
    arc.style.stroke = color;
    arc.style.strokeDashoffset = CIRC * (1 - score / 100);
  }

  function countUp(el, target, ms) {
    var t0 = performance.now();
    (function tick(now) {
      var p = Math.min((now - t0) / ms, 1);
      el.textContent = Math.round(p * target);
      if (p < 1) requestAnimationFrame(tick);
    }(performance.now()));
  }

  function animateBar(barId, pts, max, delay, color) {
    var row   = document.getElementById(barId);
    var fill  = row.querySelector('.breakdown-fill');
    var label = row.querySelector('.breakdown-pts');
    fill.style.width = '0%';
    fill.style.background = color;
    label.textContent = '0/' + max;
    setTimeout(function () {
      fill.style.width = ((pts / max) * 100) + '%';
      var t0 = performance.now();
      (function tick(now) {
        var p = Math.min((now - t0) / 600, 1);
        label.textContent = Math.round(p * pts) + '/' + max;
        if (p < 1) requestAnimationFrame(tick);
      }(performance.now()));
    }, delay);
  }

  /* ── Main submit handler ── */
  form.addEventListener('submit', function (e) {
    e.preventDefault();

    var srcVal = document.getElementById('leadSource').value;
    var engVal = document.getElementById('engagementHistory').value;
    var budVal = document.getElementById('budget').value;
    var timVal = document.getElementById('timing').value;
    var msgVal = document.getElementById('leadMessage').value;

    if (!srcVal || !engVal || !budVal || !timVal) return;

    /* ── Show processing state ── */
    idle.style.display        = 'none';
    resultPanel.style.display = 'none';
    processing.style.display  = 'flex';
    btnLabel.textContent      = 'Processing...';

    var procMessages = [
      'Scanning message...',
      'Detecting behavioural signals...',
      'Calculating lead score...',
      'Determining routing...',
    ];
    var mi = 0;
    var msgTimer = setInterval(function () {
      mi++;
      if (mi < procMessages.length) procLabel.textContent = procMessages[mi];
    }, 350);

    setTimeout(function () {
      clearInterval(msgTimer);

      /* ── SCORE CALCULATION ── */

      // Cat 1: Source (0-20)
      var sSource = SRC_SCORE[srcVal] || 0;

      // Cat 2: Engagement velocity (0-20)
      var sEngagement = ENG_SCORE[engVal] || 0;

      // Cat 3: Readiness — base from budget, boosted by keywords
      var sReadiness = BUDGET_BASE[budVal] || 0;

      var hotDetected  = [];
      var coldDetected = [];
      var advisories   = [];

      HOT_KEYWORDS.forEach(function (kw) {
        if (detect(msgVal, kw.terms)) {
          hotDetected.push(kw);
          if (kw.advisory) advisories.push(kw.advisory);
          if (kw.cat === 'readiness') sReadiness = cap(sReadiness + kw.pts, 20);
        }
      });

      // Cat 4: Intent (0-20, pure keyword)
      var sIntent = 0;
      hotDetected.filter(function (k) { return k.cat === 'intent'; }).forEach(function (k) {
        sIntent = cap(sIntent + k.pts, 20);
      });

      // Cat 5: SA Resilience (0-20, pure keyword)
      var sResilience = 0;
      hotDetected.filter(function (k) { return k.cat === 'resilience'; }).forEach(function (k) {
        sResilience = cap(sResilience + k.pts, 20);
      });

      // Cold keyword penalty (capped at -8, applied to intent only)
      var coldPenalty = 0;
      COLD_KEYWORDS.forEach(function (kw) {
        if (detect(msgVal, kw.terms)) {
          coldDetected.push(kw);
          coldPenalty = Math.min(coldPenalty + 4, 8);
        }
      });
      sIntent = Math.max(sIntent - coldPenalty, 0);

      // Timing modifier
      var timMod   = TIMING_MOD[timVal] || 1.0;
      var rawTotal = sSource + sEngagement + sReadiness + sIntent + sResilience;
      var total    = Math.round(rawTotal * timMod);

      // Priority tier
      var priority = PRIORITY.find(function (p) { return total >= p.min; });

      var hasMsgContent = msgVal.trim().length > 10;
      var signalCount   = hotDetected.length;
      var confidence    = signalCount >= 3 ? 'High confidence'
                        : signalCount >= 1 ? 'Moderate confidence'
                        : hasMsgContent    ? 'Low confidence — no clear signals'
                                           : 'Add message text to increase accuracy';

      /* ── RENDER ── */
      processing.style.display  = 'none';
      resultPanel.style.display = 'flex';
      btnLabel.textContent      = 'Re-score';

      animateRing(total, priority.color);
      countUp(scoreNum, total, 1000);

      badge.textContent       = priority.label;
      badge.style.background  = priority.color;
      badge.style.color       = priority.textColor;
      badge.style.borderColor = priority.color;
      routingLine.textContent = priority.agent;

      confidenceLine.textContent = confidence;
      confidenceLine.style.color = signalCount >= 3 ? priority.color : 'var(--color-text-muted)';

      // Signal chips
      signalsRow.innerHTML = '';
      if (hotDetected.length || coldDetected.length) {
        signalsWrap.style.display = 'block';
        hotDetected.forEach(function (kw) {
          var chip = document.createElement('span');
          chip.className   = 'signal-chip signal-hot';
          chip.textContent = kw.label;
          signalsRow.appendChild(chip);
        });
        coldDetected.forEach(function (kw) {
          var chip = document.createElement('span');
          chip.className   = 'signal-chip signal-cold';
          chip.textContent = kw.label;
          signalsRow.appendChild(chip);
        });
      } else {
        signalsWrap.style.display = 'none';
      }

      // Breakdown bars (staggered)
      animateBar('bar-source',     sSource,     20, 100,  priority.color);
      animateBar('bar-engagement', sEngagement, 20, 180,  priority.color);
      animateBar('bar-readiness',  sReadiness,  20, 260,  priority.color);
      animateBar('bar-intent',     sIntent,     20, 340,  priority.color);
      animateBar('bar-resilience', sResilience, 20, 420,  priority.color);

      // Routing card
      rChannel.textContent = priority.channel;
      rAgent.textContent   = priority.agent;
      rSLA.textContent     = priority.sla;
      rSLA.style.color     = priority.color;

      // Contextual advisories
      advisoriesEl.innerHTML = '';
      if (advisories.length) {
        var divider = document.createElement('div');
        divider.style.cssText = 'height:1px;background:var(--color-border);margin:6px 0 10px;';
        advisoriesEl.appendChild(divider);

        var header = document.createElement('div');
        header.style.cssText = 'font-family:var(--font-mono);font-size:0.62rem;' +
          'letter-spacing:0.12em;text-transform:uppercase;color:var(--color-text-muted);margin-bottom:8px;';
        header.textContent = 'Contextual Advisories';
        advisoriesEl.appendChild(header);

        advisories.forEach(function (adv) {
          var row = document.createElement('div');
          row.className = 'routing-row advisory-row';
          row.innerHTML =
            '<span class="routing-key" style="color:' + priority.color + ';">→</span>' +
            '<span class="routing-val" style="font-size:0.78rem;">' + adv + '</span>';
          advisoriesEl.appendChild(row);
        });
      }

    }, 1600);
  });

}());


/* ----------------------------------------------------------------
   2. CODE BLOCK — COPY BUTTON
---------------------------------------------------------------- */
(function () {

  var btn     = document.getElementById('copyCodeBtn');
  var snippet = document.getElementById('codeSnippet');

  if (!btn || !snippet) return;

  btn.addEventListener('click', function () {
    navigator.clipboard.writeText(snippet.innerText).then(function () {
      btn.textContent = 'Copied!';
      btn.style.color = 'var(--accent)';
      setTimeout(function () {
        btn.textContent = 'Copy';
        btn.style.color = '';
      }, 2000);
    });
  });

}());
