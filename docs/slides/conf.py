project = 'slides'
copyright = '2023, Алексей Юдолевич'
author = 'Алексей Юдолевич'

extensions = [
    'myst_parser',
    'sphinx_revealjs',
]

templates_path = ['_templates']
exclude_patterns = [
    '_build', 'Thumbs.db', '.DS_Store',
    'custom_html.md', '*_topic_*.md',
]

language = 'ru'

revealjs_style_theme = 'league'
revealjs_script_plugins = [
    {
        "name": "RevealHighlight",
        "src": "revealjs4/plugin/highlight/highlight.js",
    },
]
revealjs_css_files = [
    "revealjs4/plugin/highlight/zenburn.css",
]
