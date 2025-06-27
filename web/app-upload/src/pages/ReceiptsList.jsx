// src/pages/ReceiptsList.jsx
import { useEffect, useState } from "react";
import axios from "axios";
import { useNavigate } from "react-router-dom";

export default function ReceiptsList() {
    const [receipts, setReceipts] = useState([]);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        axios
            .get("http://localhost:8000/receipts")
            .then((res) => {
                setReceipts(res.data.receipts || []);
            })
            .catch((err) => {
                console.error("Error fetching receipts:", err);
            })
            .finally(() => {
                setLoading(false);
            });
    }, []);

    if (loading) return <p style={{ textAlign: "center" }}>Loading receipts...</p>;

    return (
        <div style={{ maxWidth: "800px", margin: "2rem auto", padding: "1rem" }}>
            <h2 style={{ textAlign: "center", marginBottom: "2rem" }}>🧾 All Receipts</h2>

            {receipts.length === 0 ? (
                <p style={{ textAlign: "center" }}>No receipts found.</p>
            ) : (
                receipts.map((r) => (
                    <div
                        key={r.id}
                        style={{
                            border: "1px solid #ddd",
                            padding: "1rem",
                            borderRadius: "8px",
                            marginBottom: "1rem",
                            boxShadow: "0 1px 5px rgba(0,0,0,0.1)",
                            cursor: "pointer",
                        }}
                        onClick={() => navigate(`/receipts/${r.id}`)}
                    >
                        <h3 style={{ marginBottom: "0.5rem" }}>{r.merchant_name}</h3>
                        <p>🕒 {new Date(r.purchased_at).toLocaleString()}</p>
                        <p>💰 ₹{r.total_amount}</p>
                    </div>
                ))
            )}
        </div>
    );
}
