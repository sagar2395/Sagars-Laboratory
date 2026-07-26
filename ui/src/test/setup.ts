// SPDX-License-Identifier: Apache-2.0
//
// Global test setup for the component layer.
//
// The MSW server is started once for the whole run. Handlers are reset between
// tests so per-test overrides cannot leak, and unhandled requests error rather
// than silently hitting the network — a test that forgets to mock an endpoint
// must fail loudly, not pass by accident.

import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from './server'

beforeAll(() => {
  server.listen({ onUnhandledRequest: 'error' })
})

afterEach(() => {
  cleanup()
  server.resetHandlers()
})

afterAll(() => {
  server.close()
})
