import { Component, type ErrorInfo, type ReactNode } from "react";

interface AppErrorBoundaryProps {
  children: ReactNode;
}

interface AppErrorBoundaryState {
  hasError: boolean;
}

export class AppErrorBoundary extends Component<
  AppErrorBoundaryProps,
  AppErrorBoundaryState
> {
  state: AppErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(): AppErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("Frontend render error", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <main className="flex min-h-dvh items-center justify-center bg-[#f5f3ee] px-6 text-center text-slate-800">
          <div className="max-w-sm space-y-3 rounded-2xl border border-rose-200 bg-white p-6 shadow-xl">
            <h1 className="text-lg font-bold">The board crashed while rendering</h1>
            <p className="text-sm text-slate-600">
              Reload the page. If it happens again, the browser console will now show the exact frontend error.
            </p>
          </div>
        </main>
      );
    }

    return this.props.children;
  }
}