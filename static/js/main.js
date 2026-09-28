/* ================================================================
   main.js — Global scripts for all public pages
   JO4 Dev | Runs after Bootstrap + AOS CDN scripts load
================================================================ */

/* ── 1. Native Animate On Scroll (Zero-dependency IntersectionObserver) ── */
(function () {
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
  }, { threshold: 0.05, rootMargin: '0px 0px -40px 0px' });

  document.querySelectorAll('[data-aos]').forEach(function (el) {
    observer.observe(el);
  });
}());

/* ── 2. Sticky nav — add .scrolled class on scroll ── */
(function () {
  var nav = document.getElementById('siteNav');
  if (!nav) return;

  function onScroll() {
    nav.classList.toggle('scrolled', window.scrollY > 24);
  }

  window.addEventListener('scroll', onScroll, { passive: true });
  onScroll(); // run on load in case page is already scrolled
}());

/* ── 3. Mobile nav toggle ── */
(function () {
  var toggle = document.getElementById('navToggle');
  var links  = document.getElementById('navLinks');
  if (!toggle || !links) return;

  toggle.addEventListener('click', function () {
    var isOpen = links.classList.toggle('open');
    toggle.classList.toggle('open', isOpen);
    toggle.setAttribute('aria-expanded', String(isOpen));
    document.body.style.overflow = isOpen ? 'hidden' : '';
  });

  // Close menu when any nav link is clicked
  links.querySelectorAll('a').forEach(function (a) {
    a.addEventListener('click', function () {
      links.classList.remove('open');
      toggle.classList.remove('open');
      toggle.setAttribute('aria-expanded', 'false');
      document.body.style.overflow = '';
    });
  });
}());

/* ── 4. Footer — dynamic year ── */
(function () {
  var el = document.getElementById('footerYear');
  if (el) el.textContent = new Date().getFullYear();
}());
