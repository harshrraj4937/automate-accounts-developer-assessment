import { BrowserRouter , Routes, Route } from "react-router-dom";
import './App.css';
import HomePage from "./pages/HomePage";
import ReceiptsList from "./pages/ReceiptsList";
import ReceiptDetail from "./pages/ReceiptDetail";
import UploadFlow from "./pages/UploadFlow";

function App() {
  return (
    <div className="App">
      <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/receipts" element={<ReceiptsList />} />
        <Route path="/receipts/:id" element={<ReceiptDetail />} />
        <Route path="/upload" element={<UploadFlow />} />
      </Routes>
    </BrowserRouter>
    </div>
  )
}

export default App;
