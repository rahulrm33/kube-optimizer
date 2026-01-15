# Static Assets

This directory contains static assets for the web dashboard.

Currently, all assets (CSS, JavaScript, and fonts) are loaded from CDNs:
- Bootstrap 5.3.0
- Bootstrap Icons 1.11.1
- Chart.js 4.4.0

If you want to serve these assets locally for offline use, download them and place them here:

```
web/static/
├── css/
│   ├── bootstrap.min.css
│   └── bootstrap-icons.css
├── js/
│   ├── bootstrap.bundle.min.js
│   └── chart.umd.js
└── fonts/
    └── bootstrap-icons.woff2
```

Then update the template references from CDN URLs to local paths.

