import { useState } from 'react'
import PatternCard from './components/Card'

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <PatternCard id={1} />
    </>
  );
}

export default App
