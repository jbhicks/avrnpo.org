// Quill Editor initialization for AVR admin panels
// Handles rich text editing functionality

(function() {
    'use strict';
    
    // Default Quill configuration
    const defaultConfig = {
        theme: 'snow',
        modules: {
            toolbar: [
                [{ 'header': [1, 2, 3, false] }],
                ['bold', 'italic', 'underline', 'strike'],
                ['blockquote', 'code-block'],
                [{ 'list': 'ordered'}, { 'list': 'bullet' }],
                [{ 'script': 'sub'}, { 'script': 'super' }],
                [{ 'indent': '-1'}, { 'indent': '+1' }],
                [{ 'direction': 'rtl' }],
                [{ 'color': [] }, { 'background': [] }],
                [{ 'font': [] }],
                [{ 'align': [] }],
                ['clean'],
                ['link', 'image']
            ]
        },
        placeholder: 'Enter your content here...'
    };
    
    // Simple toolbar for basic editing
    const simpleConfig = {
        theme: 'snow',
        modules: {
            toolbar: [
                ['bold', 'italic', 'underline'],
                ['blockquote'],
                [{ 'list': 'ordered'}, { 'list': 'bullet' }],
                ['clean'],
                ['link']
            ]
        },
        placeholder: 'Enter your content here...'
    };
    
    // Initialize Quill editors
    function initializeQuillEditors() {
        // Initialize full-featured editors
        const fullEditors = document.querySelectorAll('.quill-editor');
        fullEditors.forEach(element => {
            if (element.dataset.quillInitialized) return;
            
            const quill = new Quill(element, defaultConfig);
            element.dataset.quillInitialized = 'true';
            
            // If there's a hidden input associated, sync content
            const hiddenInput = element.parentNode.querySelector('input[type="hidden"]');
            if (hiddenInput) {
                // Set initial content if present
                if (hiddenInput.value) {
                    quill.root.innerHTML = hiddenInput.value;
                }
                
                // Sync content on change
                quill.on('text-change', function() {
                    hiddenInput.value = quill.root.innerHTML;
                });
            }
        });
        
        // Initialize simple editors
        const simpleEditors = document.querySelectorAll('.quill-editor-simple');
        simpleEditors.forEach(element => {
            if (element.dataset.quillInitialized) return;
            
            const quill = new Quill(element, simpleConfig);
            element.dataset.quillInitialized = 'true';
            
            // If there's a hidden input associated, sync content
            const hiddenInput = element.parentNode.querySelector('input[type="hidden"]');
            if (hiddenInput) {
                // Set initial content if present
                if (hiddenInput.value) {
                    quill.root.innerHTML = hiddenInput.value;
                }
                
                // Sync content on change
                quill.on('text-change', function() {
                    hiddenInput.value = quill.root.innerHTML;
                });
            }
        });
    }
    
    // Initialize when DOM is ready
    function initialize() {
        // Check if Quill is available
        if (typeof Quill === 'undefined') {
            console.warn('Quill editor not available');
            return;
        }
        
        initializeQuillEditors();
        
        // Re-initialize on HTMX content swaps
        if (typeof htmx !== 'undefined') {
            htmx.onLoad(function(content) {
                if (content.querySelectorAll) {
                    initializeQuillEditors();
                }
            });
        }
    }
    
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initialize);
    } else {
        initialize();
    }
    
    // Make Quill utilities available globally
    window.QuillManager = {
        initialize: initialize,
        initializeQuillEditors: initializeQuillEditors,
        defaultConfig: defaultConfig,
        simpleConfig: simpleConfig
    };
})();
