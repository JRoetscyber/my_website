/* ================================================================
   main.js — Global scripts for all public pages
   JO4 Dev | Runs after Bootstrap + AOS CDN scripts load
================================================================ */

/* ── 1. AOS — Animate On Scroll ── */
if (typeof AOS !== 'undefined') {
  AOS.init({
    duration: 640,
    easing:   'ease-out-cubic',
    once:     true,
    offset:   60,
  });
}

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
