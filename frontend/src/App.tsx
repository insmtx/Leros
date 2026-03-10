import { useState } from 'react';

import viteLogo from '/vite.svg';
import reactLogo from './assets/react.svg';

function App() {
  const [count, setCount] = useState(0);

  return (
    <>
      <div className="mx-auto flex justify-center p-2 text-center">
        <a href="https://vitejs.dev" target="_blank" rel="noreferrer">
          <img
            src={viteLogo}
            className="h-40 p-7 transition-all hover:drop-shadow-2xl"
            alt="Vite logo"
          />
        </a>
        <a href="https://react.dev" target="_blank" rel="noreferrer">
          <img
            src={reactLogo}
            className="h-40 animate-spin p-7 transition-all hover:drop-shadow-2xl"
            alt="React logo"
          />
        </a>
      </div>
      <h1 className="mx-auto py-10 text-center font-bold text-6xl text-gray-700">
        Vite + React
      </h1>
      <div className="flex flex-col items-center p-2">
        <button
          type="button"
          className="rounded-lg bg-slate-100 px-4 py-2 font-semibold outline-none ring-blue-500 transition-all focus:ring-2 hover:ring-1"
          onClick={() => setCount((count) => count + 1)}
        >
          count is {count}
        </button>
        <p className="mt-4">
          Edit <code>src/App.tsx</code> and save to test HMR
        </p>
      </div>
      <p className="mt-12 text-center text-slate-500">
        Click on the Vite and React logos to learn more
      </p>
    </>
  );
}

export default App;
