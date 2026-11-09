import BusTracker from './components/Bustracker.jsx'

export default function App() {
  return (
    <div className="h-screen w-screen">
      <BusTracker busId={1} busNumber="WB-001" />
    </div>
  )
}