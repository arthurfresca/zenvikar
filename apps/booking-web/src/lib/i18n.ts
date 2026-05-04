export type Locale = "en" | "pt";

export type BookingTranslations = {
  workspace_not_found: string;
  workspace_not_found_desc: string;
  sign_in_to_book: string;
  booking_success: string;
  booking_error: string;
  upcoming_visits: string;
  total: string;
  choose_specialist: string;
  change_specialist: string;
  no_services: string;
  duration_min: string;
  select_specialist: string;
  choose_provider: string;
  pick_date: string;
  check: string;
  available_times: string;
  no_times: string;
  try_different_day: string;
  sign_in: string;
  sign_in_to_confirm: string;
  ends: string;
  status_confirmed: string;
  status_cancelled: string;
  status_pending: string;
  welcome_back: string;
  sign_in_desc: string;
  email_address: string;
  email_placeholder: string;
  password: string;
  password_placeholder: string;
  signing_in: string;
  or_continue: string;
  new_here: string;
  create_account_link: string;
  create_account: string;
  create_account_desc: string;
  full_name: string;
  name_placeholder: string;
  preferred_language: string;
  creating_account: string;
  already_have_account: string;
  sign_in_link: string;
  sign_out: string;
  confirm_booking: string;
  back_to_times: string;
  no_upcoming_visits: string;
  no_upcoming_visits_desc: string;
  view_all_bookings: string;
  specialist_label: string;
  service_label: string;
  time_label: string;
  price_label: string;
  error_slot_taken: string;
  error_server: string;
};

const en: BookingTranslations = {
  workspace_not_found: "Workspace not found",
  workspace_not_found_desc: "Could not resolve a booking page for:",
  sign_in_to_book: "Sign in to book",
  booking_success: "Your appointment has been booked successfully. We look forward to seeing you!",
  booking_error: "We could not complete the booking. The time may no longer be available — please try again.",
  upcoming_visits: "Your upcoming visits",
  total: "total",
  choose_specialist: "Choose a specialist",
  change_specialist: "Choose a different specialist",
  no_services: "No services are available right now.",
  duration_min: "min",
  select_specialist: "Select a specialist",
  choose_provider: "Choose a provider to see their available times.",
  pick_date: "Pick a date",
  check: "Check",
  available_times: "Available times",
  no_times: "No available times on this date.",
  try_different_day: "Try selecting a different day.",
  sign_in: "Sign in",
  sign_in_to_confirm: "to confirm a booking",
  ends: "ends",
  status_confirmed: "Confirmed",
  status_cancelled: "Cancelled",
  status_pending: "Pending",
  welcome_back: "Welcome back",
  sign_in_desc: "Sign in to access your bookings",
  email_address: "Email address",
  email_placeholder: "you@example.com",
  password: "Password",
  password_placeholder: "••••••••",
  signing_in: "Signing in…",
  or_continue: "or continue with",
  new_here: "New here?",
  create_account_link: "Create an account",
  create_account: "Create your account",
  create_account_desc: "Start booking appointments in seconds",
  full_name: "Full name",
  name_placeholder: "Jane Smith",
  preferred_language: "Preferred language",
  creating_account: "Creating account…",
  already_have_account: "Already have an account?",
  sign_in_link: "Sign in",
  sign_out: "Sign out",
  confirm_booking: "Confirm your booking",
  back_to_times: "Back to times",
  no_upcoming_visits: "No upcoming visits",
  no_upcoming_visits_desc: "Book your first appointment below!",
  view_all_bookings: "View all",
  specialist_label: "Specialist",
  service_label: "Service",
  time_label: "Time",
  price_label: "Price",
  error_slot_taken: "This time slot is no longer available. Please choose another time.",
  error_server: "A server error occurred. Please try again later.",
};

const pt: BookingTranslations = {
  workspace_not_found: "Área de trabalho não encontrada",
  workspace_not_found_desc: "Não foi possível encontrar uma página de agendamento para:",
  sign_in_to_book: "Entrar para agendar",
  booking_success: "Seu agendamento foi confirmado com sucesso. Estamos ansiosos para recebê-lo!",
  booking_error: "Não foi possível concluir o agendamento. O horário pode não estar mais disponível — tente novamente.",
  upcoming_visits: "Suas próximas visitas",
  total: "total",
  choose_specialist: "Escolha um especialista",
  change_specialist: "Escolha um especialista diferente",
  no_services: "Nenhum serviço disponível no momento.",
  duration_min: "min",
  select_specialist: "Selecione um especialista",
  choose_provider: "Escolha um profissional à esquerda para ver os horários disponíveis.",
  pick_date: "Escolha uma data",
  check: "Verificar",
  available_times: "Horários disponíveis",
  no_times: "Nenhum horário disponível nesta data.",
  try_different_day: "Tente selecionar outro dia.",
  sign_in: "Entrar",
  sign_in_to_confirm: "para confirmar um agendamento",
  ends: "termina",
  status_confirmed: "Confirmado",
  status_cancelled: "Cancelado",
  status_pending: "Pendente",
  welcome_back: "Bem-vindo de volta",
  sign_in_desc: "Entre para acessar seus agendamentos",
  email_address: "Endereço de e-mail",
  email_placeholder: "voce@exemplo.com",
  password: "Senha",
  password_placeholder: "••••••••",
  signing_in: "Entrando…",
  or_continue: "ou continue com",
  new_here: "Novo aqui?",
  create_account_link: "Criar uma conta",
  create_account: "Crie sua conta",
  create_account_desc: "Comece a agendar compromissos em segundos",
  full_name: "Nome completo",
  name_placeholder: "Maria Silva",
  preferred_language: "Idioma preferido",
  creating_account: "Criando conta…",
  already_have_account: "Já tem uma conta?",
  sign_in_link: "Entrar",
  sign_out: "Sair",
  confirm_booking: "Confirmar agendamento",
  back_to_times: "Voltar para horários",
  no_upcoming_visits: "Nenhuma visita próxima",
  no_upcoming_visits_desc: "Agende seu primeiro compromisso abaixo!",
  view_all_bookings: "Ver todos",
  specialist_label: "Especialista",
  service_label: "Serviço",
  time_label: "Horário",
  price_label: "Preço",
  error_slot_taken: "Este horário não está mais disponível. Por favor, escolha outro horário.",
  error_server: "Ocorreu um erro no servidor. Por favor, tente novamente mais tarde.",
};

export function getTranslations(locale: string): BookingTranslations {
  return locale === "pt" ? pt : en;
}
