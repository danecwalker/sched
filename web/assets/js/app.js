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
})();