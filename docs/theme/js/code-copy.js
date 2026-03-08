(function () {
  function addCopyButtons() {
    var blocks = document.querySelectorAll('div.highlight, article pre');
    blocks.forEach(function (block) {
      if (block.closest('.code-block-wrapper')) return;
      if (block.tagName === 'PRE' && block.closest('div.highlight')) return;
      var wrapper = document.createElement('div');
      wrapper.className = 'code-block-wrapper';
      block.parentNode.insertBefore(wrapper, block);
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'copy-code-btn';
      btn.textContent = 'Copy';
      btn.setAttribute('aria-label', 'Copy code');
      var code = block.querySelector('pre code');
      var text = code ? code.textContent : (block.querySelector('pre') && block.querySelector('pre').textContent) || '';
      btn.onclick = function () {
        navigator.clipboard.writeText(text).then(function () {
          btn.textContent = 'Copied!';
          btn.classList.add('copied');
          setTimeout(function () {
            btn.textContent = 'Copy';
            btn.classList.remove('copied');
          }, 1500);
        });
      };
      wrapper.appendChild(btn);
      wrapper.appendChild(block);
    });
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', addCopyButtons);
  } else {
    addCopyButtons();
  }
})();
