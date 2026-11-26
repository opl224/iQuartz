import { useState } from "react";
import { createOrder } from "../api/orders";

const OrderForm = () => {
  const [customer, setCustomer] = useState("");
  const [sku, setSku] = useState("");
  const [qty, setQty] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<any>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const res = await createOrder({
        customer,
        items: [{ sku: sku.toUpperCase(), qty: Number(qty) }],
      });
      setSuccess(res.data);
      setSku("");
      setQty(1);
    } catch (err: any) {
      if (err.response?.data?.error?.code === "INSUFFICIENT_STOCK") {
        const { sku: failedSku, available } = err.response.data.error;
        setError(`Stok ${failedSku} tidak cukup. Tersedia: ${available}`);
      } else {
        setError(err.response?.data?.error || "Terjadi kesalahan.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        border: "1px solid #e5e7eb",
        padding: "2rem",
        borderRadius: "4px",
        marginTop: "2rem",
        maxWidth: "600px",
      }}
    >
      <h3
        style={{
          fontSize: "1.1rem",
          fontWeight: "600",
          marginBottom: "1.5rem",
          borderBottom: "1px solid #f3f4f6",
          paddingBottom: "0.5rem",
        }}
      >
        Buat Pesanan Baru
      </h3>

      <form
        onSubmit={handleSubmit}
        style={{ display: "flex", flexDirection: "column", gap: "1rem" }}
      >
        <div>
          <label
            style={{
              display: "block",
              fontSize: "0.85rem",
              marginBottom: "0.3rem",
              color: "#4b5563",
            }}
          >
            Nama Customer
          </label>
          <input
            className="input-field"
            value={customer}
            onChange={(e) => setCustomer(e.target.value)}
            required
          />
        </div>

        <div style={{ display: "flex", gap: "1rem" }}>
          <div style={{ flex: 1 }}>
            <label
              style={{
                display: "block",
                fontSize: "0.85rem",
                marginBottom: "0.3rem",
                color: "#4b5563",
              }}
            >
              SKU Produk
            </label>
            <input
              className="input-field"
              value={sku}
              onChange={(e) => setSku(e.target.value)}
              required
            />
          </div>
          <div style={{ width: "80px" }}>
            <label
              style={{
                display: "block",
                fontSize: "0.85rem",
                marginBottom: "0.3rem",
                color: "#4b5563",
              }}
            >
              Qty
            </label>
            <input
              className="input-field"
              type="number"
              min="1"
              value={qty}
              onChange={(e) => setQty(Number(e.target.value))}
              required
            />
          </div>
        </div>

        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? "Memproses..." : "Konfirmasi Pesanan"}
        </button>
      </form>

      {error && (
        <div className="msg-error" style={{ marginTop: "1rem" }}>
          {error}
        </div>
      )}
      {success && (
        <div className="msg-success" style={{ marginTop: "1rem" }}>
          Sukses! ID: {success.id} | Total: Rp{" "}
          {success.total.toLocaleString("id-ID")}
        </div>
      )}
    </div>
  );
};

export default OrderForm;
