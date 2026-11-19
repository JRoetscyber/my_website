/* ================================================================
   main.js — Global scripts for JO4 Dev (Pure Vanilla + HTMX Hooks)
   Zero external dependencies • Mobile-First • Accessible
================================================================ */

(function () {
  'use strict';

  /* ── 1. Native Animate On Scroll (IntersectionObserver) ──────── */
  function initAOS() {
    if (!('IntersectionObserver' in window)) {
      document.querySelectorAll('[data-aos]').forEach(function (el) {
        el.classList.add('aos-animate');
      });
      return;
    }

    var observer = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (entry.isIntersecting) {
          var delay = entry.target.getAttribute('data-aos-delay');
          if (delay) {
            entry.target.style.transitionDelay = delay + 'ms';
          }
          entry.target.classList.add('aos-animate');
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.05, rootMargin: '0px 0px -30px 0px' });

    document.querySelectorAll('[data-aos]:not(.aos-animate)').forEach(function (el) {
      observer.observe(el);
    });
  }

  /* ── 2. Sticky Nav Polish (rAF debounce) ─────────────────────── */
  function initStickyNav() {
    var nav = document.getElementById('siteNav');
    if (!nav) return;

    var ticking = false;
    function updateNav() {
      nav.classList.toggle('scrolled', window.scrollY > 16);
      ticking = false;
    }

    window.addEventListener('scroll', function () {
      if (!ticking) {
        window.requestAnimationFrame(updateNav);
        ticking = true;
      }
    }, { passive: true });

    updateNav();
  }

  /* ── 3. Accessible Mobile Nav Drawer ────────────────────────── */
  function initMobileNav() {
    var toggle = document.getElementById('navToggle');
    var links  = document.getElementById('navLinks');
    if (!toggle || !links) return;

    function closeMenu() {
      links.classList.remove('open');
      toggle.classList.remove('open');
      toggle.setAttribute('aria-expanded', 'false');
      document.body.style.overflow = '';
    }

    function toggleMenu() {
      var isOpen = links.classList.toggle('open');
      toggle.classList.toggle('open', isOpen);
      toggle.setAttribute('aria-expanded', String(isOpen));
      document.body.style.overflow = isOpen ? 'hidden' : '';
    }

    // Toggle click/tap
    toggle.addEventListener('click', function (e) {
      e.stopPropagation();
      toggleMenu();
    });

    // Close when clicking any nav link
    links.querySelectorAll('a').forEach(function (a) {
      a.addEventListener('click', closeMenu);
    });

    // Close on outside tap/click
    document.addEventListener('click', function (e) {
      if (links.classList.contains('open') && !links.contains(e.target) && !toggle.contains(e.target)) {
        closeMenu();
      }
    });

    // Close on Escape key
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && links.classList.contains('open')) {
        closeMenu();
      }
    });

    // Reset overflow on desktop resize
    window.addEventListener('resize', function () {
      if (window.innerWidth >= 1024 && links.classList.contains('open')) {
        closeMenu();
      }
    }, { passive: true });
  }

  /* ── 4. Dynamic Year in Footer ───────────────────────────────── */
  function initYear() {
    var el = document.getElementById('footerYear');
    if (el) el.textContent = new Date().getFullYear();
  }

  /* ── 5. HTMX Lifecycle Integration ──────────────────────────── */
  function initHTMX() {
    document.addEventListener('htmx:afterSwap', function () {
      initAOS();
      initStickyNav();
      initMobileNav();
      initYear();
      // Scroll to top on page swap if not hashtag
      if (!window.location.hash) {
        window.scrollTo({ top: 0, behavior: 'instant' });
      }
    });
  }

  /* Initialize on DOM Ready */
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () {
      initAOS();
      initStickyNav();
      initMobileNav();
      initYear();
      initHTMX();
    });
  } else {
    initAOS();
    initStickyNav();
    initMobileNav();
    initYear();
    initHTMX();
  }

})();
