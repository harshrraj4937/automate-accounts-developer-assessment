// src/pages/ReceiptDetail.jsx
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import axios from "axios";

export default function ReceiptDetail() {
    const { id } = useParams();
    const [receipt, setReceipt] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        axios
            .get(`http://localhost:8000/receipts/${id}`)
            .then((res) => {
                setReceipt(res.data);
            })
            .catch((err) => {
                console.error("Error fetching receipt:", err);
            })
            .finally(() => {
                setLoading(false);
            });
    }, [id]);

    if (loading) return <p style={{ textAlign: "center" }}>Loading...</p>;
    if (!receipt) return <p style={{ textAlign: "center" }}>Receipt not found.</p>;

    return (
        <div style={{ maxWidth: "600px", margin: "2rem auto", padding: "1rem" }}>
            <h2 style={{ textAlign: "center", marginBottom: "2rem" }}>🧾 Receipt Details</h2>

            <div
                style={{
                    border: "1px solid #ccc",
                    borderRadius: "8px",
                    padding: "1.5rem",
                    boxShadow: "0 2px 8px rgba(0, 0, 0, 0.05)",
                    backgroundColor: "#f9f9f9",
                    textAlign: "left", 
                }}
            >
                <p><strong>ID:</strong> {receipt.id}</p>
                <p><strong>Merchant:</strong> {receipt.merchant_name}</p>
                <p><strong>Purchased At:</strong> {new Date(receipt.purchased_at).toLocaleString()}</p>
                <p><strong>Total Amount:</strong> ₹{receipt.total_amount}</p>
                <p>
                    <strong>Uploaded File:</strong>{" "}
                    <a
                        href={`http://localhost:8000/${receipt.file_path}`}
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        {receipt.file_path}
                    </a>
                </p>
                <p><strong>Created At:</strong> {new Date(receipt.created_at).toLocaleString()}</p>
                <p><strong>Updated At:</strong> {new Date(receipt.updated_at).toLocaleString()}</p>
            </div>
        </div>
    );
}
