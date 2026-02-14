import { Outlet } from 'react-router-dom'

function App() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-50 to-slate-100">
      <main className="mx-auto w-full max-w-4xl px-4 py-10 sm:px-6 lg:py-14">
        <div className="rounded-2xl border border-slate-200/80 bg-white p-6 shadow-sm sm:p-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}

export default App
