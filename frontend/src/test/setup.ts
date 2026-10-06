import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";

// Vitest runs without globals, so Testing Library cannot auto-cleanup.
afterEach(() => {
  cleanup();
});
