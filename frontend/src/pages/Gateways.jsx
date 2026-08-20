import { useState } from 'react';
import { useStore } from '../data/store';
import { PageHead, StatusBadge, Empty, Field } from '../components/ui';
import Modal from '../components/Modal';
import { Confirm } from '../components/ui';
import Pagination from '../components/Pagination';

const blank = { name: '', status: 'Active' };

export default function Gateways() {
  const { db, add, update, remove } = useStore();
  const [editing, setEditing] = useState(null); // record or 'new'
  const [form, setForm] = useState(blank);
  const [confirmId, setConfirmId] = useState(null);

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage, setItemsPerPage] = useState(25);

  const openNew = () => { setForm(blank); setEditing('new'); };
  const openEdit = (g) => { setForm({ name: g.name, status: g.status }); setEditing(g); };

  const save = () => {
    if (!form.name.trim()) return;
    if (editing === 'new') add('gateways', form);
    else update('gateways', editing.id, form);
    setEditing(null);
  };

  // Paginated data
  const totalGateways = db.gateways.length;
  const startIndex = (currentPage - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  const paginatedGateways = db.gateways.slice(startIndex, endIndex);

  const handlePageChange = (page) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleItemsPerPageChange = (newItemsPerPage) => {
    setItemsPerPage(newItemsPerPage);
    setCurrentPage(1);
  };

  return (
    <div>
      <PageHead
        title="Payment Gateway Master"
        sub="Create and manage payment gateways available across the system."
        actions={<button className="btn primary" onClick={openNew}>+ Add Gateway</button>}
      />

      <div className="panel">
        {db.gateways.length === 0 ? (
          <div className="panel-body"><Empty icon="⬢" text="No gateways added yet." /></div>
        ) : (
          <>
            <table className="data">
              <thead>
                <tr>
                  <th>Gateway Name</th>
                  <th>Assigned To (Companies)</th>
                  <th>Status</th>
                  <th className="center">Actions</th>
                </tr>
              </thead>
              <tbody>
                {paginatedGateways.map((g) => {
                  const usedBy = db.companies.filter((c) => (c.gateways || []).some((x) => x.gatewayId === g.id)).length;
                  return (
                    <tr key={g.id}>
                      <td className="bold">{g.name}</td>
                      <td>{usedBy} compan{usedBy === 1 ? 'y' : 'ies'}</td>
                      <td><StatusBadge status={g.status} /></td>
                      <td className="center">
                        <div className="row-actions" style={{ justifyContent: 'center' }}>
                          <button className="btn sm ghost" onClick={() => openEdit(g)}>Edit</button>
                          <button className="btn sm danger" onClick={() => setConfirmId(g.id)}>Delete</button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {totalGateways > 10 && (
              <Pagination
                currentPage={currentPage}
                totalItems={totalGateways}
                itemsPerPage={itemsPerPage}
                onPageChange={handlePageChange}
                onItemsPerPageChange={handleItemsPerPageChange}
              />
            )}
          </>
        )}
      </div>

      {editing && (
        <Modal
          title={editing === 'new' ? 'Add Payment Gateway' : 'Edit Payment Gateway'}
          onClose={() => setEditing(null)}
          footer={<>
            <button className="btn ghost" onClick={() => setEditing(null)}>Cancel</button>
            <button className="btn primary" onClick={save}>Save</button>
          </>}
        >
          <Field label="Gateway Name">
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="e.g. Paytm" autoFocus />
          </Field>
          <Field label="Status">
            <select value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
              <option>Active</option>
              <option>Inactive</option>
            </select>
          </Field>
        </Modal>
      )}

      {confirmId && (
        <Confirm
          message="Delete this gateway? This cannot be undone."
          onCancel={() => setConfirmId(null)}
          onConfirm={() => { remove('gateways', confirmId); setConfirmId(null); }}
        />
      )}
    </div>
  );
}
