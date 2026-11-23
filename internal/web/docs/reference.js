Scalar.createApiReference('#reference', {
  url: '/openapi.yaml',
  servers: [{ url: window.location.origin }],
  withDefaultFonts: false,
  telemetry: false,
  agent: { disabled: true },
  showDeveloperTools: 'never',
  hideClientButton: true,
});
