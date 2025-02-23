(function(){
  const installButton = document.getElementById('install');
  const installText = installButton.querySelector('span');
  const installLoading = installButton.querySelector('svg');
  let deferredPrompt;

  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault();
    deferredPrompt = e;
    installButton.disabled = false;
    installText.style.display = 'block';
    installLoading.style.display = 'none';
  });

  installButton.addEventListener('click', (e) => {
    installButton.disabled = true;
    installText.style.display = 'none';
    installLoading.style.display = 'block';
    deferredPrompt.prompt();
    deferredPrompt.userChoice.then((choiceResult) => {
      if (choiceResult.outcome === 'accepted') {
        console.log('User accepted the A2HS prompt');
      } else {
        console.log('User dismissed the A2HS prompt');
      }
      deferredPrompt = null;
    });
  });


  window.addEventListener('appinstalled', (event) => {
    console.log('PWA installed successfully');
    // Navigate to a specific page upon installation
    window.location.href = '/'; 
  });
  
})();