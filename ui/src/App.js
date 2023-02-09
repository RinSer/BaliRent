import './App.css';
import 'leaflet/dist/leaflet.css';
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from './components/Login';
import Property from './components/Property';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login/>} />
        <Route path="/property" element={<Property/>} />
      </Routes>
    </BrowserRouter>
  );
}
