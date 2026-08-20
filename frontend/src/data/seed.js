// Seed data for client demo. Persisted to localStorage on first load.

export const seedData = {
  gateways: [
    { id: 'gw1', name: 'Paytm', status: 'Active' },
    { id: 'gw2', name: 'PayU', status: 'Active' },
    { id: 'gw3', name: 'Razorpay', status: 'Active' },
    { id: 'gw4', name: 'Cashfree', status: 'Active' },
    { id: 'gw5', name: 'PhonePe', status: 'Active' },
    { id: 'gw6', name: 'CCAvenue', status: 'Inactive' },
  ],

  companies: [
    {
      id: 'co1',
      name: 'Skyline Retail Pvt Ltd',
      contactPerson: 'Rahul Mehta',
      contactNumber: '9876543210',
      whatsapp: '9876543210',
      telegram: '@skyline_rahul',
      email: 'rahul@skylineretail.com',
      altContactPerson: 'Neha Sharma',
      altContactNumber: '9811122233',
      address: '14 MG Road, Bengaluru, Karnataka 560001',
      status: 'Active',
      userId: 'skyline',
      password: 'skyline123',
      gateways: [
        { gatewayId: 'gw1', commission: 4, chargeBearer: 'Admin' },
        { gatewayId: 'gw2', commission: 7, chargeBearer: 'Company' },
        { gatewayId: 'gw3', commission: 6, chargeBearer: 'Company' },
      ],
    },
    {
      id: 'co2',
      name: 'Orbit Digital Solutions',
      contactPerson: 'Anjali Verma',
      contactNumber: '9900112233',
      whatsapp: '9900112233',
      telegram: '@orbit_anjali',
      email: 'anjali@orbitdigital.in',
      altContactPerson: '',
      altContactNumber: '',
      address: '88 Anna Salai, Chennai, Tamil Nadu 600002',
      status: 'Active',
      userId: 'orbit',
      password: 'orbit123',
      gateways: [
        { gatewayId: 'gw4', commission: 5, chargeBearer: 'Admin' },
        { gatewayId: 'gw5', commission: 5.5, chargeBearer: 'Company' },
      ],
    },
  ],

  affiliates: [
    {
      id: 'af1',
      name: 'Prime Partners',
      contact: '9090909090',
      altContact: '8080808080',
      email: 'contact@primepartners.in',
      commissionPct: 1.5,
      commissionBase: 'Settlement Amount',
      userId: 'prime',
      password: 'prime123',
      status: 'Active',
    },
  ],

  merchants: [
    {
      id: 'me1',
      name: 'FashionHub Store',
      contact: '9123456780',
      altContact: '9123000000',
      email: 'support@fashionhub.in',
      companyId: 'co1',
      affiliateId: null,
      commissionPct: 1,
      commissionBase: 'Settlement Amount',
      userId: 'fashionhub',
      password: 'fashion123',
      status: 'Active',
      banks: [
        {
          id: 'bk1',
          bankName: 'HDFC Bank',
          accountName: 'FashionHub Store',
          accountNumber: '50100123456789',
          ifsc: 'HDFC0001234',
          netbankingLink: 'https://netbanking.hdfcbank.com',
          username: 'fashionhub_net',
          loginPassword: 'Net@1234',
          txnPassword: 'Txn@5678',
          customerId: 'CUST889900',
          mobile: '9123456780',
          email: 'finance@fashionhub.in',
          mobileBanking: 'Yes',
          mobileLoginId: 'fashionhub_mob',
          mpin: '4567',
          atmCards: [
            { id: 'atm1', nameOnCard: 'FASHIONHUB STORE', cardNumber: '4111 1111 1111 1111', expiry: '08/28', cvv: '123', atmPin: '4321' },
          ],
          custom: [],
        },
      ],
      paymentGateways: [
        {
          id: 'pg1', gatewayId: 'gw1', loginLink: 'https://dashboard.paytm.com',
          merchantId: 'PAYTM-FH-001', username: 'fashionhub', password: 'Paytm@123',
          mobile: '9123456780', email: 'finance@fashionhub.in', custom: [],
        },
      ],
    },
    {
      id: 'me2',
      name: 'GadgetWorld',
      contact: '9555666777',
      altContact: '',
      email: 'hello@gadgetworld.in',
      companyId: 'co1',
      affiliateId: 'af1',
      commissionPct: 0,
      commissionBase: 'Settlement Amount',
      userId: 'gadgetworld',
      password: 'gadget123',
      status: 'Active',
      banks: [
        {
          id: 'bk2',
          bankName: 'ICICI Bank',
          accountName: 'GadgetWorld',
          accountNumber: '602101987654',
          ifsc: 'ICIC0006021',
          netbankingLink: 'https://infinity.icicibank.com',
          username: 'gadget_world',
          loginPassword: 'Gw@9999',
          txnPassword: 'Tx@1111',
          customerId: 'IC772211',
          mobile: '9555666777',
          email: 'accounts@gadgetworld.in',
        },
      ],
    },
    {
      id: 'me3',
      name: 'HomeDecor Plus',
      contact: '9333222111',
      altContact: '',
      email: 'care@homedecorplus.in',
      companyId: 'co2',
      affiliateId: null,
      commissionPct: 1.2,
      commissionBase: 'Transaction Amount',
      userId: 'homedecor',
      password: 'home123',
      status: 'Active',
      banks: [],
    },
  ],

  transactions: [
    {
      id: 'tx1', companyId: 'co1', merchantId: 'me1', gatewayId: 'gw1',
      date: '2026-06-01', txnAmount: 50000, settlementAmount: 49500,
      txnCharges: 500, otherCharges: 0, remarks: 'Order #10231',
    },
    {
      id: 'tx2', companyId: 'co1', merchantId: 'me1', gatewayId: 'gw2',
      date: '2026-06-03', txnAmount: 120000, settlementAmount: 118500,
      txnCharges: 1500, otherCharges: 200, remarks: 'Bulk order',
    },
    {
      id: 'tx3', companyId: 'co1', merchantId: 'me2', gatewayId: 'gw3',
      date: '2026-06-05', txnAmount: 80000, settlementAmount: 79000,
      txnCharges: 1000, otherCharges: 0, remarks: 'Electronics sale',
    },
    {
      id: 'tx4', companyId: 'co2', merchantId: 'me3', gatewayId: 'gw4',
      date: '2026-06-07', txnAmount: 65000, settlementAmount: 64200,
      txnCharges: 800, otherCharges: 100, remarks: 'Furniture',
    },
  ],

  settlements: [
    {
      id: 'st1', companyId: 'co1', date: '2026-06-10', amount: 40000,
      paymentMode: 'NEFT', refNumber: 'NEFT202606100001', remarks: 'Part settlement',
    },
    {
      id: 'st2', companyId: 'co2', date: '2026-06-12', amount: 30000,
      paymentMode: 'UPI', refNumber: 'UPI778899', remarks: '',
    },
  ],

  affiliatePayments: [
    { id: 'ap1', affiliateId: 'af1', date: '2026-06-11', amount: 500, paymentMode: 'UPI', refNumber: 'UPI112233', remarks: '' },
  ],

  merchantPayments: [
    { id: 'mp1', merchantId: 'me1', date: '2026-06-11', amount: 1000, paymentMode: 'Bank Transfer', refNumber: 'TRF99887', remarks: '' },
  ],
};

export const PAYMENT_MODES = ['Bank Transfer', 'UPI', 'Cash', 'Cheque', 'IMPS', 'NEFT', 'RTGS', 'Other'];
export const CHARGE_BEARERS = ['Admin', 'Company'];
export const COMMISSION_BASES = ['Transaction Amount', 'Settlement Amount'];
