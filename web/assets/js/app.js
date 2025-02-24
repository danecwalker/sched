import { Gradient } from './gradient.js';

function isPWA() {
  return window.matchMedia('(display-mode: standalone)').matches ||
         window.navigator.standalone === true;
}

(function() {
  if (!isPWA()) {
    window.location.href = "/install";
  }


  // Create your instance
  const gradient = new Gradient()
  // Call `initGradient` with the selector to your canvas
  gradient.initGradient('#gradient-canvas')

  let refresh = document.querySelector('#refresh');
  refresh.addEventListener('click', function() {
    window.location.reload();
  });

  let totalHours = document.querySelector('#total-hours');
  let totalPay = document.querySelector('#total-pay');

  let flip = true;

  totalHours.addEventListener('click', function() {
    if (!flip) {
      return;
    }
    flip = !flip;
    totalHours.style.display = 'none';
    totalPay.style.display = 'flex';
  });

  totalPay.addEventListener('click', function() {
    if (flip) {
      return;
    }
    flip = !flip;
    totalHours.style.display = 'flex';
    totalPay.style.display = 'none';
  });
})();