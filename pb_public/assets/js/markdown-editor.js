const easyMDE = new EasyMDE({
  element: document.getElementById('markdown-editor'),
  spellChecker: false,
  minHeight: '400px',
  placeholder: 'Write your post content in Markdown...',
  status: ['lines', 'words', 'cursor'],
  toolbar: [
    'bold',
    'italic',
    'strikethrough',
    '|',
    'heading-1',
    'heading-2',
    'heading-3',
    '|',
    'quote',
    'unordered-list',
    'ordered-list',
    '|',
    'link',
    'image',
    'code',
    'table',
    '|',
    'preview',
    'side-by-side',
    'fullscreen',
    '|',
    'guide'
  ]
});

document.querySelector('form').addEventListener('submit', function(e) {
  const content = easyMDE.value();
  if (!content || content.trim() === '') {
    e.preventDefault();
    alert('Please enter post content');
    easyMDE.codemirror.focus();
    return false;
  }
});
