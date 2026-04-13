import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { StackView } from './pages/StackView.jsx'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/stack/default" replace />} />
        <Route path="/stack/:stackName" element={<StackView />} />
        <Route path="/stack/:stackName/:nodeId" element={<StackView />} />
      </Routes>
    </BrowserRouter>
  )
}
